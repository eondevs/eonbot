package outcome

import (
	"eonbot/pkg/exchange"
	"errors"

	"github.com/shopspring/decimal"
)

const (
	CalcCounterPercent = "counterpercent"
	CalcCounterUnits   = "counterunits"
	CalcBasePercent    = "basepercent"
	CalcBaseUnits      = "baseunits"
)

type Buy struct {
	Price  string          `json:"price" conform:"trim,lower"`
	Amount decimal.Decimal `json:"amount"`

	// Calc specifies how Amount field should be interpreted.
	// If units calc is used, amount field means counter asset amount to spend.
	// If percent calc is used, amount field means current counter asset percent
	// amount to spend.
	Calc string `json:"calcType" conform:"trim,lower"`

	// BasePercent specifies whether base percent calc type is allowed or not.
	BasePercent bool
}

func (b Buy) Validate() error {
	switch b.Price {
	case exchange.LastPrice, exchange.AskPrice, exchange.BidPrice:
		break
	default:
		return errors.New("price type is invalid")
	}

	switch b.Calc {
	case CalcCounterPercent, CalcCounterUnits, CalcBaseUnits:
		break
	case CalcBasePercent:
		if !b.BasePercent {
			return errors.New("calc type is invalid")
		}
		break
	default:
		return errors.New("calc type is invalid")
	}

	if b.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be a possitive value")
	}

	return nil
}

func (b *Buy) Reset() {}

func (b Buy) PrepAmount(baseBal, counterBal, baseRate decimal.Decimal) (decimal.Decimal, error) {
	switch b.Calc {
	case CalcCounterPercent:
		val := counterBal.Mul(b.Amount.Div(decimal.New(100, 0)))
		if val.GreaterThan(counterBal) {
			return decimal.Zero, errors.New("specified base asset amount to buy exceeds counter asset balance")
		}
		return val.Div(baseRate), nil
	case CalcCounterUnits:
		if b.Amount.GreaterThan(counterBal) {
			return decimal.Zero, errors.New("specified base asset amount to buy exceeds counter asset balance")
		}
		return b.Amount.Div(baseRate), nil
	case CalcBasePercent:
		if !b.BasePercent {
			return decimal.Zero, errors.New("calc type is invalid")
		}
		val := baseBal.Mul(b.Amount.Div(decimal.New(100, 0)))
		if val.Mul(baseRate).GreaterThan(counterBal) {
			return decimal.Zero, errors.New("specified base asset amount to buy exceeds counter asset balance")
		}
		return val, nil
	case CalcBaseUnits:
		if b.Amount.Mul(baseRate).GreaterThan(counterBal) {
			return decimal.Zero, errors.New("specified base asset amount to buy exceeds counter asset balance")
		}
		return b.Amount, nil
	default:
		return decimal.Zero, errors.New("calc type is invalid")
	}
}
