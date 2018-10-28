package bot

import (
	"eonbot/pkg"
	"eonbot/pkg/asset"
	"eonbot/pkg/config"
	"eonbot/pkg/remote/inner"
	"eonbot/pkg/strategy"
	"eonbot/pkg/stream"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

/*
   streams setup
*/

// reconfigureStreams creates new streams and/or updates
// them if changes in main/sub/strategies configs were made.
// Singular configs' check time must be updated outside of this
// function.
func (b *botProcess) reconfigureStreams() error {
	// no changes in main/sub/strategies configs were made, return.
	if !b.Conf.Strategies().IsChanged() && !b.Conf.MainConfig().IsChanged() && !b.Conf.SubConfigs().IsChanged() {
		return nil
	}

	// remove streams whose pairs
	// are not 'active' in the main
	// config anymore.
	b.streamsCleanUp()

	// keep a list of all stream pairs, whose strategies
	// need to be checked and their check time updated.
	// If true is set, stream strategies will be
	// updated if change will be found and check time update,
	// if false is set, only strategies check time will be updated.
	uncheckedStratPairs := make(map[string]bool, 0)

	// check if main/sub configs were changed.
	if b.Conf.MainConfig().IsChanged() || b.Conf.SubConfigs().IsChanged() {
		// keep a list of sub configs' names, whose check time
		// needs to be updated.
		checkedSubs := make(map[string]string)

		// loop over all active pairs in the main config.
		for _, pair := range b.Conf.MainConfig().Get().BotConfig.ActivePairs {
			// check if pair's stream exists.
			strm, exists := b.streams[pair.String()]

			if exists { // if pair's stream does exist, update its config/strategies (if needed).
				// retrieve sub config's name for check time updating
				// and usage in stream.
				subName := b.Conf.SubConfigs().GetName(pair)
				if subName != "" {
					checkedSubs[subName] = ""
				}

				// first check for sub config and only if it's
				// not found, use main config.
				// check if sub config changed, was removed, etc.
				switch b.Conf.SubConfigs().IsSubChanged(pair) {
				case config.Changed: // if sub config was just added or changed.
					// retrieve specific sub config.
					sub, _ := b.Conf.SubConfigs().Get(pair)

					// gather strategies by pair.
					strats, err := b.gatherStreamStrategies(pair, subName, sub.Strategies)
					if err != nil {
						return err
					}

					// create new stream and add it to the
					// streams map.
					newStr, err := stream.New(
						pair,
						stream.StreamConfig{
							Config: sub,
							IsMain: false,
						}, b.RC, b.DB, b.Exchange, strats)
					if err != nil {
						return err
					}
					b.streams[pair.String()] = newStr
					uncheckedStratPairs[pair.String()] = false
				case config.NotExists: // sub config does not exist or was just removed
					// check if main config was used by this stream and updated or sub config removed.
					if strm.Conf.IsMain && b.Conf.MainConfig().IsChanged() || !strm.Conf.IsMain {
						// retrieve main config.
						main := b.Conf.MainConfig().Get()

						// gather strategies by pair.
						strats, err := b.gatherStreamStrategies(pair, "main", main.PairsConfig.Strategies)
						if err != nil {
							return err
						}

						// create new stream and add it to the
						// streams map.
						newStr, err := stream.New(
							pair,
							stream.StreamConfig{
								Config: main.PairsConfig,
								IsMain: true,
							}, b.RC, b.DB, b.Exchange, strats)
						if err != nil {
							return err
						}
						b.streams[pair.String()] = newStr
						uncheckedStratPairs[pair.String()] = false
						break
					}
					fallthrough
				case config.NotChanged:
					uncheckedStratPairs[pair.String()] = true
				}
			} else { // if pair's stream does not exist, create it.
				// retrieve pair's config.
				conf, isMain := b.Conf.PairConfig(pair)
				confName := "main"

				// check if main config was used or not.
				if !isMain {
					// retrieve sub config's name.
					confName = b.Conf.SubConfigs().GetName(pair)
					checkedSubs[confName] = ""
				}

				// gather strategies by pair.
				strats, err := b.gatherStreamStrategies(pair, confName, conf.Strategies)
				if err != nil {
					return err
				}

				// create new stream and add it to the
				// streams map.
				newStr, err := stream.New(
					pair,
					stream.StreamConfig{
						Config: conf,
						IsMain: isMain,
					}, b.RC, b.DB, b.Exchange, strats)
				if err != nil {
					return err
				}

				b.streams[pair.String()] = newStr
				uncheckedStratPairs[pair.String()] = false
			}
		}

		// update check time *after* updating streams because the
		// same sub config could be used more than once.
		if len(checkedSubs) > 0 {
			for sub := range checkedSubs {
				// update sub config check time by its name.
				b.Conf.SubConfigs().UpdateSubCheckTime(sub)
			}
		}
	} else {
		for _, pair := range b.Conf.MainConfig().Get().BotConfig.ActivePairs {
			uncheckedStratPairs[pair.String()] = true
		}
	}

	// check if any of the strategies were changed.
	if b.Conf.Strategies().IsChanged() {
		// keep a list of strategies' names, whose check time
		// needs to be updated.
		checkedStrats := make(map[string]string)

		// check if pairs with unchecked strategies
		// exist.
		if len(uncheckedStratPairs) > 0 {
			for pair, check := range uncheckedStratPairs {
				// retrieve unchecked pair's stream.
				strm, exists := b.streams[pair]
				if !exists {
					continue
				}

				// loop over unchecked pair's stream's strategies
				// and check if they were changed.
				for k, v := range strm.StrategiesInUse() {
					if check {
						switch b.Conf.Strategies().IsStratChanged(k) {
						case config.Changed: // if strategy was changed, update it.
							strm.UpdateStrategy(v, b.Conf.Strategies().Get(k))
						case config.NotExists: // if strategy does not exist, but is being used by the stream, return error.
							return fmt.Errorf("'%s' strategy, which is being used by %s pair, does not exist", k, strm.Pair.String())
						}
					}

					checkedStrats[k] = ""
				}
			}
		}

		// update check time *after* updating streams because the
		// same strategy could be used more than once.
		if len(checkedStrats) > 0 {
			for strat := range checkedStrats {
				// update strategy check time by its file name.
				b.Conf.Strategies().UpdateStratCheckTime(strat)
			}
		}
	}

	return nil
}

// streamsCleanUp removes streams of pairs that
// are not being used anymore (inactive in main config).
func (b *botProcess) streamsCleanUp() {
	// check if main config was changed.
	if b.Conf.MainConfig().IsChanged() {
		// check if at least one stream exists.
		if len(b.streams) > 0 {
		Streams:
			for k, strm := range b.streams {
				// loop over active pairs and check
				// if stream's pair is active.
				for _, pair := range b.Conf.MainConfig().Get().BotConfig.ActivePairs {
					if strm.Pair.Equal(pair) {
						continue Streams
					}
				}

				// remove pair's stream from
				// the streams map.
				delete(b.streams, k)
			}
		}
	}
}

// gatherStreamStrategies gathers strategies used
// by specific asset pair.
func (b *botProcess) gatherStreamStrategies(pair asset.Pair, confName string, confStrategies []string) ([]strategy.Strategy, error) {
	res := make([]strategy.Strategy, 0)

	// loop over config's strategies names.
Outer:
	for _, str := range confStrategies {
		for _, strat := range b.Conf.Strategies().GetAll() {
			// if strategy exists, add it to the slice.
			if strat.Name() == str {
				res = append(res, strat)
				continue Outer
			}
		}

		// return error if strategy does not exist.
		return nil, fmt.Errorf("'%s' strategy specified in %s config does not exist", str, confName)
	}

	// if pair doesn't have any strategies in its config,
	// return error.
	if len(res) <= 0 {
		return nil, fmt.Errorf("%s config does not have any strategies specified", confName)
	}

	return res, nil
}

/*
   normal execution
*/

// execNormal checks if cooldown is active or not,
// collects balances and starts stream in a normal mode.
func (b *botProcess) execNormal() {
	if len(b.streams) == 0 {
		logrus.WithField("action", "normal cycle execution").Error("initialized streams list cannot be empty")
		return
	}

	logrus.StandardLogger().Debug("starting cycle execution")
	// retrieve cooldown info.
	cooldown, err := b.Exchange.GetCooldownInfo()
	if err != nil {
		logrus.WithField("action", "normal cycle cooldown info retrieval").Error(err)
		return
	}

	// if cooldown is active, stop
	// function execution and return.
	if cooldown.Active {
		logrus.StandardLogger().Debug("exchange cooldown is active, stopping cycle exection")
		return
	}

	// check if cooldown was activated after/during functions execution.
	defer func() {
		// retrieve cooldown info.
		cooldown, err := b.Exchange.GetCooldownInfo()
		if err != nil {
			logrus.WithField("action", "completed normal cycle cooldown info retrieval").Error(err)
			return
		}

		if cooldown.Active {
			// notify RC about cooldown activation.
			b.RC.InternalSend(inner.CooldownActivationEvent)
		}
	}()

	// prepare utils that will be used
	// to wait for all streams to complete their
	// execution.
	var wg sync.WaitGroup
	wg.Add(len(b.streams))

	// limiter is used to limit how many concurrent
	// streams should be running at the same time.
	limiter := make(chan bool, b.Conf.MainConfig().Get().BotConfig.StreamCount)

	// retrieve balances of all pairs.
	balances, err := b.Exchange.GetBalances()
	if err != nil {
		logrus.WithField("action", "normal cycle balances retrieval").Error(err)
		return
	}

	// loop over streams map and start
	// them in normal mode.
	for _, s := range b.streams {
		// prepare asset pair's balances.
		var bal stream.BalancesPair

		// find base asset balance.
		if base, ok := balances[string(s.Pair.Base)]; ok {
			bal.Base = base
		}

		// find counter asset balance.
		if counter, ok := balances[string(s.Pair.Counter)]; ok {
			bal.Counter = counter
		}

		// try to send value to limiter channel,
		// if channel is full (max amount of streams
		// are running), wait until one of the streams
		// finishes.
		limiter <- true

		logrus.StandardLogger().Debugf("starting %s execution", s.Pair.String())

		// start stream's normal mode execution in
		// a separate goroutine.
		go func(started time.Time, strm *stream.Stream) {
			// exec stream in normal mode.
			res, err := strm.Normal(bal)

			if err != nil {
				logrus.StandardLogger().Error(err)
			}

			// group cycle result data.
			cyc := pkg.NewStreamCycle(started, time.Now().UTC(), res, err)

			// save pair's cycle info to db.
			if err := b.DB.Persistent().SavePairCycle(strm.Pair, cyc); err != nil {
				logrus.WithField("action", "cycle saving to db").Error(err)
			}
			wg.Done()
			<-limiter
			logrus.StandardLogger().Debugf("completed %s execution", strm.Pair.String())
		}(time.Now().UTC(), s)
	}

	// wait until all streams complete.
	wg.Wait()

	// notify RC about cycle end.
	b.RC.InternalSend(inner.PairCycleEndEvent)

	logrus.StandardLogger().Debug("completed cycle execution")
}

/*
   side tasks execution
*/

// execSellAll checks if cooldown is active or not,
// collects balances and sells all base assets.
// Returns false if execution was not successful.
func (b *botProcess) execSellAll() bool {
	// retrieve cooldown info.
	cooldown, err := b.Exchange.GetCooldownInfo()
	if err != nil {
		logrus.WithField("action", "sell side task cooldown info retrieval").Error(err)
		return false
	}

	// if cooldown is active, stop
	// function execution and return.
	if cooldown.Active {
		return true
	}

	// prepare utils that will be used
	// to wait for all streams to complete their
	// execution.
	var wg sync.WaitGroup
	wg.Add(len(b.streams))

	// limiter is used to limit how many concurrent
	// streams should be running at the same time.
	limiter := make(chan bool, b.Conf.MainConfig().Get().BotConfig.StreamCount)

	// retrieve balances of all pairs.
	balances, err := b.Exchange.GetBalances()
	if err != nil {
		logrus.WithField("action", "sell side task balances retrieval").Error(err)
		return false
	}

	var mu sync.Mutex
	success := true
	for _, s := range b.streams {
		// prepare asset pair's balances.
		var bal stream.BalancesPair

		// find base asset balance.
		if base, ok := balances[string(s.Pair.Base)]; ok {
			bal.Base = base
		}

		// find counter asset balance.
		if counter, ok := balances[string(s.Pair.Counter)]; ok {
			bal.Counter = counter
		}

		// try to send value to limiter channel,
		// if channel is full (max amount of streams
		// are running), wait until one of the streams
		// finishes.
		limiter <- true
		go func(strm *stream.Stream) {
			// execute stream in sell mode.
			if err := strm.Sell(bal); err != nil {
				mu.Lock()
				if success {
					success = false
				}
				mu.Unlock()
				logrus.WithField("action", fmt.Sprintf("%s pair sell side task execution", strm.Pair.String())).Error(err)
			}
			wg.Done()
			<-limiter
		}(s)
	}

	// wait until all streams complete.
	wg.Wait()

	return success
}

// execCancelAll checks if cooldown is active or not
// and cancels all open orders.
// Returns false if execution was not successful.
func (b *botProcess) execCancelAll() bool {
	// retrieve cooldown info.
	cooldown, err := b.Exchange.GetCooldownInfo()
	if err != nil {
		logrus.WithField("action", "cancel all side task cooldown info retrieval").Error(err)
		return false
	}

	// if cooldown is active, stop
	// function execution and return.
	if cooldown.Active {
		return true
	}

	// prepare utils that will be used
	// to wait for all streams to complete their
	// execution.
	var wg sync.WaitGroup
	wg.Add(len(b.streams))

	// limiter is used to limit how many concurrent
	// streams should be running at the same time.
	limiter := make(chan bool, b.Conf.MainConfig().Get().BotConfig.StreamCount)

	var mu sync.Mutex
	success := true
	for _, s := range b.streams {
		// try to send value to limiter channel,
		// if channel is full (max amount of streams
		// are running), wait until one of the streams
		// finishes.
		limiter <- true
		go func(strm *stream.Stream) {
			// execute stream in cancel mode.
			if err := strm.CancelAll(); err != nil {
				mu.Lock()
				if success {
					success = false
				}
				mu.Unlock()
				logrus.WithField("action", fmt.Sprintf("%s pair cancel all side task execution", strm.Pair.String())).Error(err)
			}
			wg.Done()
			<-limiter
		}(s)
	}

	// wait until all streams complete.
	wg.Wait()

	return success
}
