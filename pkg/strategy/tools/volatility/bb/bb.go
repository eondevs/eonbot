package bb

import (
	"eonbot/pkg/exchange"
	indiBB "eonbot/pkg/strategy/indicators/bb"
	"eonbot/pkg/strategy/tools"
	"errors"

	"github.com/shopspring/decimal"
)

const (
	bandLower = "lower"
	bandUpper = "upper"
)

type BB struct {
	bb       indiBB.BB
	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	Band string `json:"band" conform:"trim,lower"`

	indiBB.BBConfig
	tools.Cond
	tools.CondObject
	tools.Shift
}

type snapshot struct {
	ShiftedBand decimal.Decimal `json:"shiftedBand"`
	tools.CondObjectSnapshot
	indiBB.BBInfo
}

func New(conf func(v interface{}) error) (*BB, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	bb, err := indiBB.NewFromConfig(s.BBConfig, 0)
	if err != nil {
		return nil, err
	}

	s.CondObject.AllowTickerPrice()
	if err := s.CondObject.Init(0); err != nil {
		return nil, err
	}

	return &BB{
		bb:   bb,
		conf: s,
	}, nil
}

func (b *BB) Validate() error {
	switch b.conf.Band {
	case bandLower, bandUpper:
		break
	default:
		return errors.New("band type is invalid")
	}

	if err := b.conf.BBConfig.Validate(); err != nil {
		return err
	}

	if err := b.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := b.conf.CondObject.Validate(); err != nil {
		return err
	}

	if err := b.conf.Shift.Validate(); err != nil {
		return err
	}

	return nil
}

func (b *BB) ConditionsMet(d exchange.Data) (bool, error) {
	bbInfo, err := b.bb.Calc(d.Candles)
	if err != nil {
		b.snapshot.Clear()
		return false, err
	}

	var band decimal.Decimal
	switch b.conf.Band {
	case bandUpper:
		band = bbInfo.Upper
	case bandLower:
		band = bbInfo.Lower
	default:
		b.snapshot.Clear()
		return false, errors.New("band type is invalid")
	}

	val, err := b.conf.CondObject.Value(d)
	if err != nil {
		b.snapshot.Clear()
		return false, err
	}

	shiftedBand := b.conf.Shift.CalcVal(band)
	isMet := b.conf.Cond.Match(val, shiftedBand)

	b.snapshot.Set(snapshot{
		ShiftedBand:        shiftedBand,
		CondObjectSnapshot: b.conf.CondObject.Snapshot(val),
		BBInfo:             bbInfo,
	}, isMet)

	return isMet, nil
}

func (b *BB) CandlesCount() int {
	return b.bb.CandlesCount()
}

func (b *BB) Snapshot() tools.Snapshot {
	return b.snapshot.Get()
}

func (b *BB) Reset() {}
