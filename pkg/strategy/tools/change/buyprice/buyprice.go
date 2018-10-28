package buyprice

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"github.com/shopspring/decimal"
)

type BuyPrice struct {
	conf settings

	snapshot tools.SnapshotManager
}

type settings struct {
	tools.CondObject
	tools.Cond
	tools.Shift
}

type snapshot struct {
	BuyPrice        decimal.Decimal `json:"buyPrice"`
	ShiftedBuyPrice decimal.Decimal `json:"shiftedBuyPrice"`
	tools.CondObjectSnapshot
}

func New(conf func(v interface{}) error) (*BuyPrice, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	s.CondObject.AllowTickerPrice()
	if err := s.CondObject.Init(0); err != nil {
		return nil, err
	}

	return &BuyPrice{
		conf: s,
	}, nil
}

func (p *BuyPrice) Validate() error {
	if err := p.conf.CondObject.Validate(); err != nil {
		return err
	}

	if err := p.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := p.conf.Shift.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *BuyPrice) ConditionsMet(d exchange.Data) (bool, error) {
	val, err := p.conf.CondObject.Value(d)
	if err != nil {
		p.snapshot.Clear()
		return false, err
	}

	if d.BuyPrice.Equal(decimal.Zero) {
		p.snapshot.Set(snapshot{
			BuyPrice:           d.BuyPrice,
			CondObjectSnapshot: p.conf.CondObject.Snapshot(val),
		}, true)
		return true, nil
	}

	shiftedPrice := p.conf.Shift.CalcVal(d.BuyPrice)
	isMet := p.conf.Cond.Match(val, shiftedPrice)
	p.snapshot.Set(snapshot{
		BuyPrice:           d.BuyPrice,
		ShiftedBuyPrice:    shiftedPrice,
		CondObjectSnapshot: p.conf.CondObject.Snapshot(val),
	}, isMet)
	return isMet, nil
}

func (p *BuyPrice) CandlesCount() int {
	return p.conf.CondObject.CandlesCount()
}

func (p *BuyPrice) Snapshot() tools.Snapshot {
	return p.snapshot.Get()
}

func (p *BuyPrice) Reset() {}
