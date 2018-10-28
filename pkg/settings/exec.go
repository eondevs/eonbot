package settings

import "errors"

// Exec contains settings specified at the time of
// execution i.e. specified as command arguments/flags.
// It won't ever be converted to/from JSON.
type Exec struct {
	// Verbose specifies whether the bot
	// should print debug level logs
	// or not.
	Verbose bool

	// Stdout specifies whether the bot should print
	// logs info to stdout or not (true = print).
	Stdout bool

	// LogsDir specifies a filesystem path
	// to a directory where all logs files should
	// be placed.
	LogsDir string

	// LogsJSON specifies whether the logs should be
	// formatted to JSON or not (true = formatted).
	LogsJSON bool

	// ConfigsDir specifies a filesystem path
	// to a directory where all config files should
	// be placed.
	ConfigsDir string

	// SubsDir specifies a filesystem path
	// to a directory where all sub configs files should
	// be placed.
	SubsDir string

	// StrategiesDir specifies a filesystem path
	// to a directory where all strategies files should
	// be placed.
	StrategiesDir string

	// CancelAll specifies whether the bot should cancel
	// all open orders. Disabled when AutoStart is set to false.
	CancelAll bool

	// SellAll specifies whether the bot should
	// sell all base assets currently in account
	// or not. Disabled when AutoStart is set to false.
	SellAll bool

	// ReloadStop specifies whether the bot
	// should be stopped when reloaded config
	// has error. It will be logged into file and
	// not loaded if not.
	ReloadStop bool

	// AutoStart specifies whether the bot
	// should start immediately or wait for
	// RC to start it.
	AutoStart bool

	// RCPort specifies which port should be used for
	// internal remote controller.
	RCPort int

	// HTTPTimeout specifies the max amount of time the request should
	// take to go to the exchange driver and back. In seconds.
	HTTPTimeout int64
}

func (e Exec) Validate() error {
	if e.LogsDir == "" {
		return errors.New("logs directory not specified")
	}

	if e.ConfigsDir == "" {
		return errors.New("configs directory not specified")
	}

	if e.SubsDir == "" {
		return errors.New("sub configs directory not specified")
	}

	if e.StrategiesDir == "" {
		return errors.New("strategies directory not specified")
	}

	if e.RCPort <= 1024 {
		return errors.New("remote control port cannot be 1024 or less")
	}

	if e.HTTPTimeout < 10 || e.HTTPTimeout > 120 {
		return errors.New("HTTP timeout should be between 10 and 120 seconds (inclusively)")
	}
	return nil
}
