package rsi

import (
	"eonbot/pkg/exchange"
	indiRSI "eonbot/pkg/strategy/indicators/rsi"
	"eonbot/pkg/strategy/tools"

	"github.com/shopspring/decimal"
)

type RSI struct {
	rsi      indiRSI.RSI
	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	indiRSI.RSIConfig
	tools.Cond
	tools.Level
}

type snapshot struct {
	RSIVal decimal.Decimal `json:"rsiVal"`
}

func New(conf func(v interface{}) error) (*RSI, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	rsi, err := indiRSI.NewFromConfig(s.RSIConfig, 0)
	if err != nil {
		return nil, err
	}

	s.Level.ZeroToHundred()

	return &RSI{
		rsi:  rsi,
		conf: s,
	}, nil
}

func (r *RSI) Validate() error {
	if err := r.conf.RSIConfig.Validate(); err != nil {
		return err
	}

	if err := r.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := r.conf.Level.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *RSI) ConditionsMet(d exchange.Data) (bool, error) {
	rsi, err := r.rsi.Calc(d.Candles)
	if err != nil {
		r.snapshot.Clear()
		return false, err
	}

	isMet := r.conf.Cond.Match(rsi, r.conf.Level.LevelVal)

	// collect snapshot data
	r.snapshot.Set(snapshot{RSIVal: rsi}, isMet)
	return isMet, nil
}

func (r *RSI) CandlesCount() int {
	return r.rsi.CandlesCount()
}

func (r *RSI) Snapshot() tools.Snapshot {
	return r.snapshot.Get()
}

func (r *RSI) Reset() {}
