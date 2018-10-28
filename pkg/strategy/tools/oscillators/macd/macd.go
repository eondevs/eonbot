package macd

import (
	"eonbot/pkg/exchange"
	indiMACD "eonbot/pkg/strategy/indicators/macd"
	"eonbot/pkg/strategy/tools"

	"github.com/shopspring/decimal"
)

type MACD struct {
	macd     indiMACD.MACD
	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	Differ decimal.Decimal `json:"diff"`
	indiMACD.MACDConfig
	tools.Cond
	tools.Diff
}

type snapshot struct {
	Diff decimal.Decimal `json:"diffVal"`
	indiMACD.MACDInfo
}

func New(conf func(v interface{}) error) (*MACD, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	macd, err := indiMACD.NewFromConfig(s.MACDConfig, 0)
	if err != nil {
		return nil, err
	}

	return &MACD{
		macd: macd,
		conf: s,
	}, nil
}

func (m *MACD) Validate() error {
	if err := m.conf.MACDConfig.Validate(); err != nil {
		return err
	}

	if err := m.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := m.conf.Diff.Validate(); err != nil {
		return err
	}

	return nil
}

func (m *MACD) ConditionsMet(d exchange.Data) (bool, error) {
	macd, err := m.macd.Calc(d.Candles)
	if err != nil {
		m.snapshot.Clear()
		return false, err
	}

	diff := m.conf.Diff.Diff(macd.SignalLine, macd.MACDLine)
	isMet := m.conf.Cond.Match(diff, m.conf.Differ)

	// collect snapshot data
	m.snapshot.Set(snapshot{
		Diff:     diff,
		MACDInfo: macd,
	}, isMet)
	return isMet, nil
}

func (m *MACD) CandlesCount() int {
	return m.macd.CandlesCount()
}

func (m *MACD) Snapshot() tools.Snapshot {
	return m.snapshot.Get()
}

func (m *MACD) Reset() {}
