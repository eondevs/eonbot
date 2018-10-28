// Package sma implements SMA indicator calculation logic.
package sma

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"

	"github.com/shopspring/decimal"
)

type SMA interface{ ma.MA }

// sma contains internal data needed
// to calculate SMA.
type sma struct {
	period int
	offset int
	price  string
}

// New creates new sma object with provided data values
// to make further calculations.
func New(period, offset int, price string) (*sma, error) {
	if err := exchange.CandlePriceValid(price); err != nil {
		return nil, err
	}

	return &sma{
		period: period,
		offset: offset,
		price:  price,
	}, nil
}

// CandlesCount returns min candle count needed
// to calculate SMA with the provided period.
func (s *sma) CandlesCount() int {
	return s.period + s.offset
}

// Calc calculates SMA of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns SMA calculation result and optionally an error.
// 1. SMA = (val1 + val2 + ...) / valCount;
func (s *sma) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	start := s.CandlesCount()
	end := s.offset
	if cc == nil || len(cc) < start {
		return decimal.Zero, errors.New("SMA candles list is too small")
	}

	sum := decimal.Zero
	for _, c := range cc[len(cc)-start : len(cc)-end] {
		sum = sum.Add(c.Price(s.price))
	}
	return sum.Div(decimal.New(int64(s.period), 0)), nil
}

// CalcDecimal calculates SMA of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns SMA calculation result and optionally an error.
// SMA calculation:
// 1. SMA = (val1 + val2 + ...) / valCount;
func (s *sma) CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error) {
	start := s.CandlesCount()
	end := s.offset
	if dd == nil || len(dd) < start {
		return decimal.Zero, errors.New("SMA values list is too small")
	}

	sum := decimal.Zero
	for _, d := range dd[len(dd)-start : len(dd)-end] {
		sum = sum.Add(d)
	}
	return sum.Div(decimal.New(int64(s.period), 0)), nil
}
