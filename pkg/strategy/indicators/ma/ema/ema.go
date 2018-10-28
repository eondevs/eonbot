// Package ema implements EMA indicator calculation logic.
package ema

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/indicators/ma/sma"
	"errors"

	"github.com/shopspring/decimal"
)

type EMA interface{ ma.MA }

// ema contains internal values needed
// to calculate EMA.
type ema struct {
	period int
	offset int
	price  string
	sma    sma.SMA
}

// New creates new ema object with provided data value
// to make further calculations.
func New(period, offset int, price string) (*ema, error) {
	sma, err := sma.New(period, period+offset, price)
	if err != nil {
		return nil, err
	}

	return newEMA(period, offset, price, sma)
}

// newEMA creates new ema object with provided period
// and offset to make further calculations and allows
// to set a custom SMA.
func newEMA(period, offset int, price string, sma sma.SMA) (*ema, error) {
	if err := exchange.CandlePriceValid(price); err != nil {
		return nil, err
	}

	return &ema{
		period: period,
		offset: offset,
		sma:    sma,
		price:  price,
	}, nil
}

// CandlesCount returns min candle count needed
// to calculate EMA with the provided period.
func (e *ema) CandlesCount() int {
	return e.sma.CandlesCount() // SMA offset is period+offset, which results in period*2+offset
}

// Calc calculates EMA of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns EMA calculation result and optionally an error.
// EMA calculation (X represents period count):
// 1. First EMA = SMA of X period values;
// 2. Multiplier: 2 / (X + 1);
// 3. EMA = (current candle price - EMA(previous day)) x multiplier + EMA(previous day);
// 4. Repeat step 3 for X times;
func (e *ema) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	sma, err := e.sma.Calc(cc)
	if err != nil {
		// sma calc returs only one type anyway
		return decimal.Zero, errors.New("EMA candles list is too small")
	}

	start := e.CandlesCount() - e.period
	end := e.offset
	if cc == nil || len(cc) < start {
		return decimal.Zero, errors.New("EMA candles list is too small")
	}

	previousEMA := sma
	k := decimal.New(2, 0).Div(decimal.New(int64(e.period), 0).Add(decimal.New(1, 0)))
	for _, c := range cc[len(cc)-start : len(cc)-end] {
		previousEMA = c.Price(e.price).Sub(previousEMA).Mul(k).Add(previousEMA)
	}

	return previousEMA, nil
}

// CalcDecimal calculates EMA of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns EMA calculation result and optionally an error.
// EMA calculation (X represents period count):
// 1. First EMA = SMA of X period values;
// 2. Multiplier: 2 / (X + 1);
// 3. EMA = (current value - EMA(previous day)) x multiplier + EMA(previous day);
// 4. Repeat step 3 for X times;
func (e *ema) CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error) {
	sma, err := e.sma.CalcDecimal(dd)
	if err != nil {
		return decimal.Zero, err
	}

	start := e.CandlesCount() - e.period
	end := e.offset
	if dd == nil || len(dd) < start {
		return decimal.Zero, errors.New("EMA values list is too small")
	}

	previousEMA := sma
	k := decimal.New(2, 0).Div(decimal.New(int64(e.period), 0).Add(decimal.New(1, 0)))
	for _, d := range dd[len(dd)-start : len(dd)-end] {
		previousEMA = d.Sub(previousEMA).Mul(k).Add(previousEMA)
	}
	return previousEMA, nil
}
