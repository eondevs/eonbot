package simple

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"

	"github.com/shopspring/decimal"
)

type SimpleChange struct {
	val decimal.Decimal

	conf settings

	snapshot tools.SnapshotManager
}

type settings struct {
	tools.CondObject
	tools.Cond
	tools.Shift
}

type snapshot struct {
	ChangeValue decimal.Decimal `json:"changeVal"`
	tools.CondObjectSnapshot
}

func New(conf func(v interface{}) error) (*SimpleChange, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	s.CondObject.AllowTickerPrice()
	s.CondObject.AllowTickerMiscProp()
	s.CondObject.AllowCandlePrice()
	s.CondObject.AllowMA()
	if err := s.CondObject.Init(0); err != nil {
		return nil, err
	}

	s.Shift.AllowFixed()

	return &SimpleChange{
		conf: s,
	}, nil
}

func (s *SimpleChange) Validate() error {
	if err := s.conf.CondObject.Validate(); err != nil {
		return err
	}

	if err := s.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := s.conf.Shift.Validate(); err != nil {
		return err
	}

	return nil
}

func (s *SimpleChange) ConditionsMet(d exchange.Data) (isMet bool, err error) {
	val, err := s.conf.CondObject.Value(d)
	if err != nil {
		s.snapshot.Clear()
		return isMet, err
	}

	defer func() {
		// collect snapshot data
		s.snapshot.Set(snapshot{
			ChangeValue:        s.val,
			CondObjectSnapshot: s.conf.Snapshot(val),
		}, isMet)
	}()

	if s.val.Equal(decimal.Zero) {
		s.val = s.conf.Shift.CalcVal(val)
		if s.conf.Shift.Calc.Type != tools.CalcFixed {
			return isMet, nil
		}
	}

	isMet = s.conf.Cond.Match(val, s.val)
	return isMet, nil
}

func (s *SimpleChange) CandlesCount() int {
	return s.conf.CondObject.CandlesCount()
}

func (s *SimpleChange) Snapshot() tools.Snapshot {
	return s.snapshot.Get()
}

func (s *SimpleChange) Reset() {
	s.val = decimal.Zero
}
