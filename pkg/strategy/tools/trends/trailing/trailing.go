package trailing

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"

	"github.com/shopspring/decimal"
)

type TrailingTrends struct {
	leadObj tools.CondObject
	backObj tools.CondObject

	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	// BackIndex specifies how many candles behind from the latest
	// should the second candle (used to compare with the latest one) be.
	BackIndex int `json:"backIndex"`

	Differ decimal.Decimal `json:"diff"`

	tools.CondObject
	tools.Cond
	tools.Diff
}

type snapshot struct {
	Diff    decimal.Decimal `json:"diffVal"`
	LeadObj decimal.Decimal `json:"leadObjVal"`
	BackObj decimal.Decimal `json:"backObjVal"`
}

func New(conf func(v interface{}) error) (*TrailingTrends, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	s.CondObject.AllowCandlePrice()
	s.CondObject.AllowMA()

	leadObj := s.CondObject
	backObj := s.CondObject

	if err := leadObj.Init(0); err != nil {
		return nil, err
	}

	if err := backObj.Init(s.BackIndex); err != nil {
		return nil, err
	}

	return &TrailingTrends{
		leadObj: leadObj,
		backObj: backObj,
		conf:    s,
	}, nil
}

func (t *TrailingTrends) Validate() error {
	if t.conf.BackIndex < 1 || t.conf.BackIndex > 200 {
		return errors.New("back index must be between 1 and 200 (inclusively)")
	}

	if err := t.conf.CondObject.Validate(); err != nil {
		return err
	}

	if err := t.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := t.conf.Diff.Validate(); err != nil {
		return err
	}

	return nil
}

func (t *TrailingTrends) ConditionsMet(d exchange.Data) (bool, error) {
	val1, err := t.leadObj.Value(d)
	if err != nil {
		t.snapshot.Clear()
		return false, err
	}

	val2, err := t.backObj.Value(d)
	if err != nil {
		t.snapshot.Clear()
		return false, err
	}

	diff := t.conf.Diff.Diff(val2, val1)
	isMet := t.conf.Cond.Match(diff, t.conf.Differ)

	t.snapshot.Set(snapshot{
		Diff:    diff,
		LeadObj: val1,
		BackObj: val2,
	}, isMet)
	return isMet, nil
}

func (t *TrailingTrends) CandlesCount() int {
	return t.backObj.CandlesCount() // both objects are the same in length, but obj2 is X candles back
}

func (t *TrailingTrends) Snapshot() tools.Snapshot {
	return t.snapshot.Get()
}

func (t *TrailingTrends) Reset() {}
