package bot

import (
	"eonbot/pkg/config"
	"eonbot/pkg/control"
	"eonbot/pkg/db"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"eonbot/pkg/remote"
	"eonbot/pkg/settings"
	"eonbot/pkg/stream"
	"io"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Launch creates new bot process object and
// starts it.
// When called, it will only return error
// that prevents bot from running, otherwise
// it 'hangs' i.e. loops internally.
func Launch(conf settings.Exec) error {
	// validate exec config (created from
	// exec flags).
	if err := conf.Validate(); err != nil {
		return err
	}

	// configure logger used by all packages
	// of the process.
	if err := logSetup(conf); err != nil {
		return err
	}

	// create new bot process object.
	proc, err := newBotProcess(conf)
	if err != nil {
		logrus.StandardLogger().WithField("action", "initialization").Error(err)
		return nil
	}

	// start bot process.
	return proc.launch()
}

// logSetup prepares logger according to provided
// settings.
func logSetup(conf settings.Exec) error {
	// ensure that logs directory exists,
	// create it if not.
	if err := file.PrepDir(conf.LogsDir); err != nil {
		return err
	}

	// set logger format.
	if conf.LogsJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	// prepare logs file manager.
	fileLog := &lumberjack.Logger{
		Filename:   path.Join(conf.LogsDir, "eonbot.log"), // logs file location
		MaxSize:    10,                                    // max amount of megabytes
		MaxAge:     30,                                    // max amount of days for old files
		MaxBackups: 2,                                     // max amount of backups
	}

	// set output types.
	if conf.Stdout {
		logrus.SetOutput(io.MultiWriter(fileLog, os.Stdout))
	} else {
		logrus.SetOutput(fileLog)
	}

	// set which level logs should be printed.
	if conf.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	return nil
}

// botProcess ties up together different
// parts of the bot.
type botProcess struct {
	// Conf specifies configs loading,
	// updating and retrieving manager.
	Conf config.Manager

	// Control specifies workflow
	// control manager (start, stop, restart).
	Control control.Controller

	// Exchange specifies exchange driver client.
	Exchange exchange.Exchange

	// DB specifies persistent filesystem / in-memory bot store manager.
	DB db.Manager

	// RC specifies manager used to communicate with
	// external processes and services.
	RC remote.Manager

	// streams specifies asset pairs streams used
	// to execute strategies / place orders and retain
	// pair specific data between multiple cycle in an
	// encapsulated place.
	// The keys is asset pair's code.
	streams map[string]*stream.Stream
}

// newBotProcess creates new botProcess object.
func newBotProcess(conf settings.Exec) (*botProcess, error) {
	// create new pointer to botProcess.
	proc := &botProcess{}

	// create new config manager with provided exec config.
	proc.Conf = config.New(conf)

	// load configs.
	if err := proc.Conf.InitLoad(); err != nil {
		return nil, err
	}

	// create new workflow controller.
	proc.Control = control.New(proc.onStateChange)

	// create new exchange driver client.
	proc.Exchange = exchange.New(proc.Conf.ExecConfig().Get().HTTPTimeout)

	// set new exchange driver's address.
	if err := proc.Exchange.SetAddress(proc.Conf.RemoteConfig().Get().ExchangeDriverAddress); err != nil {
		return nil, err
	}

	// ping exchange driver to check if it's running and the connection
	// has no problems.
	if err := proc.Exchange.Ping(); err != nil {
		return nil, err
	}

	// create new database manager.
	dbMan, err := db.New()
	if err != nil {
		return nil, err
	}

	proc.DB = dbMan

	// create new remote control manager.
	proc.RC = remote.New(proc.Conf, proc.Control, proc.DB, proc.Exchange)

	// init streams map.
	proc.streams = make(map[string]*stream.Stream)

	return proc, nil
}

// launch starts bot's internal execution loop.
func (b *botProcess) launch() error {
	logrus.StandardLogger().Info("bot is running...")

	// init is used to check if bot finished at least one
	// cycle or not. It shows bot's process state.
	// Used for errors (i.e. should the bot exit
	// fatally or not).
	init := b.Conf.ExecConfig().Get().AutoStart

	// start termination handler in another goroutine.
	go b.handleTermination()

	// don't wait for RC start command, if
	// auto start flag is provided.
	if b.Conf.ExecConfig().Get().AutoStart {
		// exec start command.
		b.Control.Start(control.StartInfo{
			Commons: control.Commons{
				SellAll:   b.Conf.ExecConfig().Get().SellAll,
				CancelAll: b.Conf.ExecConfig().Get().CancelAll,
			},
		}, control.CauseInit)
	}

	// start bot process execution loop.
ProcessLoop:
	for {
		logrus.StandardLogger().Debug("process loop activated")

		// check if any changes were made in any of the configs
		// and try to apply them.
		if err := b.handleConfigs(); err != nil {
			logrus.StandardLogger().WithField("action", "config handling").Error(err)

			// if init state is still active,
			// stop the process.
			if init {
				return err
			}

			// if error occurs during the config applying and
			// user uses reload stop flag, stop the bot.
			if b.Conf.ExecConfig().Get().ReloadStop {
				// if start is pending, deny it.
				if b.Control.IsStartPending() {
					start := <-b.Control.WaitStart()
					start.Commons.Confirm <- false

					// since memory db reset might be omitted during the stop
					// command execution, we need to reset it here.
					b.DB.InMemory().Reset()
				}
			}
		}

		// wait for start command.
		start := <-b.Control.WaitStart()

		// confirm that start is successful and allowed.
		start.Commons.Confirm <- true

		logrus.StandardLogger().Debug("start command received")

		// execute all side tasks from the start command.
		b.execSideTasks(start.Commons)

		// create a timer for delays between cycles.
		timer := time.NewTimer(time.Nanosecond)

		// cycles loop.
	CycleLoop:
		for {
			logrus.StandardLogger().Debug("new cycle started, activating delay")

			// check if any of the configs were modified.
			mod, err := b.Conf.CheckAndLoad()
			if err != nil {
				logrus.StandardLogger().WithField("action", "config loading and parsing").Error(err)

				// if init state is still active,
				// stop the process.
				if init {
					return err
				}

				// if error occurs during the config loading/parsing and
				// user uses reload stop flag, stop the bot.
				if b.Conf.ExecConfig().Get().ReloadStop {
					// exec stop command.
					b.Control.Stop(control.StopInfo{PreventReset: true}, control.CauseConfigLoadErr)
				}
			}

			// if any of the configs were modified, restart the bot and
			// apply changes.
			if mod {
				logrus.StandardLogger().Debug("configs modified, restarting process loop")
				// drain timer channel, so that
				// only stop channel event
				// is received in select case below.
				if !timer.Stop() {
					<-timer.C
				}

				// exec restart command.
				b.Control.Restart(control.StartInfo{}, control.StopInfo{PreventReset: true}, control.CauseConfigChange)
			}

			// wait until either stop command
			// is received or delay completes.
			select {
			case stop := <-b.Control.WaitStop():

				// confirm that stop is successful and allowed.
				stop.Commons.Confirm <- true

				logrus.StandardLogger().Debug("stop command received")
				// drain timer channel.
				if !timer.Stop() {
					select {
					case <-timer.C:
						break
					case <-time.After(time.Second): // in case channel was already drained
						break
					}
				}

				// execute all side tasks from the stop command.
				b.execSideTasks(stop.Commons)

				// reset in memory db, if needed.
				if !stop.PreventReset {
					b.DB.InMemory().Reset()
				}

				// if specified, kill the whole process.
				// No remote controllers will be able to connect after this.
				if stop.Kill {
					logrus.StandardLogger().Debug("process kill command received, killing the process...")
					// stop process loop.
					break ProcessLoop
				}

				logrus.StandardLogger().Debug("process loop stopped")
				// stop cycles loop.
				break CycleLoop
			case <-timer.C:
				logrus.StandardLogger().Debug("cycle delay completed")
				b.execNormal()
			}

			// Reset timer with the specified delay duration.
			// No need for stop func call or channel draining because
			// this will already be done in the select above.
			timer.Reset(time.Duration(b.Conf.MainConfig().Get().BotConfig.CycleDelay) * time.Second)

			// change init value.
			if init {
				init = false
			}

			logrus.StandardLogger().Debug("cycle ended")
		}
	}
	return nil
}

// handleConfigs checks if specific configs were changed and
// tries to apply changes.
func (b *botProcess) handleConfigs() error {
	// check if remote config has changed.
	if b.Conf.RemoteConfig().IsChanged() {
		logrus.StandardLogger().Debug("applying remote config settings...")
		// configure telegram, if needed.
		b.RC.ConfigTelegram()

		// if exchange driver address is different than used by the exchange driver client,
		// update it.
		if b.Conf.RemoteConfig().Get().ExchangeDriverAddress != b.Exchange.GetAddress() {
			// set new exchange driver's address.
			if err := b.Exchange.SetAddress(b.Conf.RemoteConfig().Get().ExchangeDriverAddress); err != nil {
				return err
			}

			// ping exchange driver to check if it's running and the connection
			// has no problems.
			if err := b.Exchange.Ping(); err != nil {
				return err
			}

			// check/update main config's asset pairs list.
			if err := b.Conf.MainConfig().UpdateActivePairs(b.Exchange); err != nil {
				return err
			}

			// check if main config's candle interval is valid.
			if err := b.Conf.MainConfig().ValidateInterval(b.Exchange); err != nil {
				return err
			}

			// check if all sub configs have valid candle intervals
			if err := b.Conf.SubConfigs().ValidateInterval(b.Exchange); err != nil {
				return err
			}
		}
	}

	// check if main config has changed.
	if b.Conf.MainConfig().IsChanged() {
		logrus.StandardLogger().Debug("applying main config settings...")
		// check/update main config's asset pairs list.
		if err := b.Conf.MainConfig().UpdateActivePairs(b.Exchange); err != nil {
			return err
		}

		// check if main config's candle interval is valid.
		if err := b.Conf.MainConfig().ValidateInterval(b.Exchange); err != nil {
			return err
		}
	}

	// check if any of the sub configs have changed.
	if b.Conf.SubConfigs().IsChanged() {
		logrus.StandardLogger().Debug("applying sub configs settings...")
		// check if all sub configs have valid candle intervals
		if err := b.Conf.SubConfigs().ValidateInterval(b.Exchange); err != nil {
			return err
		}
	}

	// apply configs changes (if any) to the asset pairs streams.
	if err := b.reconfigureStreams(); err != nil {
		return err
	}

	// update configs check times.
	b.Conf.RemoteConfig().UpdateCheckTime()
	b.Conf.MainConfig().UpdateCheckTime()
	b.Conf.SubConfigs().UpdateCheckTime()
	b.Conf.Strategies().UpdateCheckTime()

	return nil
}

// handleTermination waits until termination signals
// are received and safely releases all open resources.
func (b *botProcess) handleTermination() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
	b.DB.CloseAll()
	b.RC.Stop()
	os.Exit(0)
}
