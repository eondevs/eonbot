// Package wma implements WMA indicator calculation logic.
package wma

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"

	"github.com/shopspring/decimal"
)

type WMA interface{ ma.MA }

// wma contains internal data values needed
// to calculate WMA.
type wma struct {
	period int
	offset int
	price  string
}

// New creates new wma object with data values
// to make further calculations.
func New(period, offset int, price string) (*wma, error) {
	if err := exchange.CandlePriceValid(price); err != nil {
		return nil, err
	}

	return &wma{
		period: period,
		offset: offset,
		price:  price,
	}, nil
}

// CandlesCount returns min candle count needed
// to calculate WMA with the provided period.
func (w *wma) CandlesCount() int {
	return w.period + w.offset
}

// Calc calculates WMA of provided period candles.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns WMA calculation result and optionally an error.
// WMA calculation:
// 0. n* starts at 1 and increments each time it's used;
// 1. WMA = (val1 * n1) + (val * n2) + ... / (total sum of all n* values);
func (w *wma) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	start := w.CandlesCount()
	end := w.offset
	if cc == nil || len(cc) < start {
		return decimal.Zero, errors.New("WMA candles list is too small")
	}

	sum := decimal.Zero
	k := decimal.New(1, 0)
	total := k
	for i, c := range cc[len(cc)-start : len(cc)-end] {
		sum = sum.Add(c.Price(w.price).Mul(k))
		if i != len(cc)-end-1 {
			k = k.Add(decimal.New(1, 0))
			total = total.Add(k)
		}
	}
	return sum.Div(total), nil
}

// CalcDecimal calculates WMA of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns WMA calculation result and optionally an error.
// 0. n* starts at 1 and increments each time it's used;
// 1. WMA = (val1 * n1) + (val * n2) + ... / (total sum of all n* values);
func (w *wma) CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error) {
	start := w.CandlesCount()
	end := w.offset
	if dd == nil || len(dd) < start {
		return decimal.Zero, errors.New("WMA values list is too small")
	}

	sum := decimal.Zero
	k := decimal.New(1, 0)
	total := k
	for i, d := range dd[len(dd)-start : len(dd)-end] {
		sum = sum.Add(d.Mul(k))
		if i != len(dd)-end-1 {
			k = k.Add(decimal.New(1, 0))
			total = total.Add(k)
		}
	}
	return sum.Div(total), nil
}
