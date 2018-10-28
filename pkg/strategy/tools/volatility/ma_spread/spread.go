package ma_spread

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators"
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/tools"
	"errors"

	"github.com/shopspring/decimal"
)

type MASpread struct {
	ma1      ma.MA
	ma2      ma.MA
	conf     settings
	snapshot tools.SnapshotManager
}

type maConf struct {
	Type string `json:"maType" conform:"trim,lower"`
	ma.MAConfig
}

type settings struct {
	Spread decimal.Decimal `json:"spread"`
	BaseMA int             `json:"baseMA"`
	MA1    maConf          `json:"ma1"`
	MA2    maConf          `json:"ma2"`
	tools.Diff
	tools.Cond
}

type snapshot struct {
	Spread decimal.Decimal `json:"spreadVal"`
	MA1    decimal.Decimal `json:"ma1Val"`
	MA2    decimal.Decimal `json:"ma2Val"`
}

func New(conf func(v interface{}) error) (*MASpread, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	ma1, err := indicators.NewMAFromConfig(s.MA1.Type, s.MA1.MAConfig, 0)
	if err != nil {
		return nil, err
	}

	ma2, err := indicators.NewMAFromConfig(s.MA2.Type, s.MA2.MAConfig, 0)
	if err != nil {
		return nil, err
	}

	return &MASpread{
		ma1:  ma1,
		ma2:  ma2,
		conf: s,
	}, nil
}

func (m *MASpread) Validate() error {
	if m.conf.BaseMA != 1 && m.conf.BaseMA != 2 {
		return errors.New("base ma index is invalid")
	}

	if err := ma.MATypeValidation(m.conf.MA1.Type); err != nil {
		return err
	}

	if err := ma.MATypeValidation(m.conf.MA2.Type); err != nil {
		return err
	}

	if err := m.conf.MA1.MAConfig.Validate(); err != nil {
		return err
	}

	if err := m.conf.MA2.MAConfig.Validate(); err != nil {
		return err
	}

	if err := ma.TwoMAValidation(m.conf.MA1.Type, m.ma1, m.conf.MA2.Type, m.ma2); err != nil {
		return err
	}

	if err := m.conf.Diff.Validate(); err != nil {
		return err
	}

	if err := m.conf.Cond.Validate(); err != nil {
		return err
	}
	return nil
}

func (m *MASpread) ConditionsMet(d exchange.Data) (bool, error) {
	maVal1, err := m.ma1.Calc(d.Candles)
	if err != nil {
		m.snapshot.Clear()
		return false, err
	}

	maVal2, err := m.ma2.Calc(d.Candles)
	if err != nil {
		m.snapshot.Clear()
		return false, err
	}

	var spread decimal.Decimal
	if m.conf.BaseMA == 1 {
		spread = m.conf.Diff.Diff(maVal1, maVal2)
	} else if m.conf.BaseMA == 2 {
		spread = m.conf.Diff.Diff(maVal2, maVal1)
	} else {
		m.snapshot.Clear()
		return false, errors.New("base ma index is invalid")
	}

	isMet := m.conf.Cond.Match(spread, m.conf.Spread)

	m.snapshot.Set(snapshot{
		Spread: spread,
		MA1:    maVal1,
		MA2:    maVal2,
	}, isMet)
	return isMet, nil
}

func (m *MASpread) CandlesCount() int {
	if m.ma1.CandlesCount() > m.ma2.CandlesCount() {
		return m.ma1.CandlesCount()
	}
	return m.ma2.CandlesCount()
}

func (m *MASpread) Snapshot() tools.Snapshot {
	return m.snapshot.Get()
}

func (m *MASpread) Reset() {}
