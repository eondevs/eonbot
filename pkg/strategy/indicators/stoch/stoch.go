// Package stoch implements Stochastic indicator calculation logic.
package stoch

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma/sma"
	"eonbot/pkg/utils"
	"errors"

	"github.com/shopspring/decimal"
)

type Stoch interface {
	CandlesCount() int
	Calc(cc []exchange.Candle) (StochInfo, error)
}

// Stoch contains internal data values
// needed to calculate StochInfo values.
type stoch struct {
	kPeriod int
	dPeriod int
	offset  int

	sma sma.SMA
}

// StochInfo contains result values of
// Stoch.Calc function.
type StochInfo struct {
	K decimal.Decimal `json:"K"`
	D decimal.Decimal `json:"D"`
}

// New creates new Stoch object with provided data values.
func New(kPeriod, dPeriod, offset int) (*stoch, error) {
	sma, err := sma.New(dPeriod, 0, exchange.ClosePrice)
	if err != nil {
		return nil, err
	}

	return newStoch(kPeriod, dPeriod, offset, sma), nil
}

// NewFromConfig creates new Stoch object the same way as New,
// it just takes values from provided StochConfig.
func NewFromConfig(conf StochConfig, offset int) (*stoch, error) {
	return New(conf.KPeriod, conf.DPeriod, offset)
}

// newStoch creates new Stoch object with specfied data values.
func newStoch(k, d, offset int, sma sma.SMA) *stoch {
	return &stoch{
		kPeriod: k,
		offset:  offset,
		dPeriod: d,
		sma:     sma,
	}
}

// CandlesCount returns min candle count needed
// to calculate Stoch with the provided period.
func (s *stoch) CandlesCount() int {
	return s.kPeriod + (s.dPeriod - 1) + s.offset
}

// Calc calculates Stoch (latest K and D) of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns Stoch calculation result and optionally an error.
// Stoch calculation (X represents period count):
// 1. %K = (Current Close - Lowest Low)/(Highest High - Lowest Low) * 100;
// 2. %D = 3-day SMA of %K (from step 1);
//
// Lowest Low = lowest low for the look-back period;
// Highest High = highest high for the look-back period;
// %K is multiplied by 100 to move the decimal point two places;
func (s *stoch) Calc(cc []exchange.Candle) (StochInfo, error) {
	start := s.CandlesCount()
	end := s.offset

	if cc == nil || len(cc) < start {
		return StochInfo{}, errors.New("Stoch candles list is too small")
	}

	candles := cc[len(cc)-start : len(cc)-end]

	var kk []decimal.Decimal
	for i := 0; i < s.dPeriod; i++ {
		lowest := decimal.Zero
		highest := decimal.Zero
		for _, candle := range candles[i : len(candles)-(s.dPeriod-1-i)] {
			if highest.Equal(decimal.Zero) || candle.High.GreaterThan(highest) {
				highest = candle.High
			}

			if lowest.Equal(decimal.Zero) || candle.Low.LessThan(lowest) {
				lowest = candle.Low
			}
		}
		close := candles[len(candles)-(s.dPeriod-i)].Close // no need for additional -1
		kk = append(kk, close.Sub(lowest).Div(utils.PreventZero(highest.Sub(lowest))).Mul(decimal.New(100, 0)))
	}

	d, err := s.sma.CalcDecimal(kk)
	if err != nil {
		return StochInfo{}, err
	}

	return StochInfo{
		K: kk[len(kk)-1],
		D: d,
	}, nil
}
