// Package bb implements BollingerBands indicator calculation logic.
package bb

import (
	"eonbot/pkg/exchange"
	ebMath "eonbot/pkg/math"
	"eonbot/pkg/strategy/indicators"
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/indicators/ma/sma"
	"errors"

	"github.com/shopspring/decimal"
)

type BB interface {
	CandlesCount() int
	Calc(cc []exchange.Candle) (BBInfo, error)
}

// bb contains internal data values
// needed to calculate BBInfo values.
type bb struct {
	period   int
	offset   int
	stdev    decimal.Decimal
	price    string
	midMA    ma.MA
	stdevSMA sma.SMA
}

// BBInfo contains result values of
// bb.Calc function.
type BBInfo struct {
	Upper  decimal.Decimal `json:"upper"`
	Middle decimal.Decimal `json:"middle"`
	Lower  decimal.Decimal `json:"lower"`
}

// New creates new bb object with provided data values.
func New(period, offset int, stdev decimal.Decimal, maType, price string) (*bb, error) {
	if err := exchange.CandlePriceValid(price); err != nil {
		return nil, err
	}

	midMA, err := indicators.NewMA(maType, price, period, offset)
	if err != nil {
		return nil, err
	}

	stdevSMA, err := sma.New(period, offset, price)
	if err != nil {
		return nil, err
	}

	return newBB(period, offset, stdev, price, midMA, stdevSMA), nil
}

// NewFromConfig creates new bb object the same way as New,
// it just takes values from provided BBConfig.
func NewFromConfig(conf BBConfig, offset int) (*bb, error) {
	return New(conf.Period, offset, conf.STDEV, conf.MAType, conf.Price)
}

// newBB creates new bb object with specfied data values.
func newBB(period, offset int, stdev decimal.Decimal, price string, midMA ma.MA, stdevSMA sma.SMA) *bb {
	return &bb{
		period:   period,
		offset:   offset,
		stdev:    stdev,
		price:    price,
		midMA:    midMA,
		stdevSMA: stdevSMA,
	}
}

// CandlesCount returns min candle count needed
// to calculate BB with the provided period.
func (b *bb) CandlesCount() int {
	return b.midMA.CandlesCount()
}

// Calc calculates BB (Middle, Upper, Lower) of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns BB calculation result and optionally an error.
// BB calculation (X represents period count):
// 0. Calculate standard deviation of X period;
// 1. MiddleBand = X period MA;
// 2. UpperBand = MiddleBand + (standard deviation x STDEV multiplier);
// 3. LowerBand = MiddleBand - (standard deviation x STDEV multiplier);
func (b *bb) Calc(cc []exchange.Candle) (BBInfo, error) {
	mid, err := b.midMA.Calc(cc)
	if err != nil {
		return BBInfo{}, err
	}

	stdDev, err := b.standardDeviation(cc)
	if err != nil {
		return BBInfo{}, err
	}

	return BBInfo{
		Upper:  mid.Add(stdDev.Mul(b.stdev)),
		Middle: mid,
		Lower:  mid.Sub(stdDev.Mul(b.stdev)),
	}, nil
}

// standardDeviation calculates standard deviation of
// provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns standard deviation calculation result and
// optionally an error.
// standard deviation calculation:
// 0. Calculate average price for the number of periods;
// 1. Determine each period's deviation (price - average price);
// 2. Square each period's deviation;
// 3. Sum the squared deviations;
// 4. Divide this sum by the period number;
// 5. The standard deviation is then equal to the square
// root of the number from step 4;
func (b *bb) standardDeviation(cc []exchange.Candle) (decimal.Decimal, error) {
	start := b.period + b.offset
	end := b.offset

	if cc == nil || len(cc) < start {
		return decimal.Zero, errors.New("BB candles list is too small")
	}

	avg, err := b.stdevSMA.Calc(cc)
	if err != nil {
		return decimal.Zero, err
	}

	sum := decimal.Zero
	for _, c := range cc[len(cc)-start : len(cc)-end] {
		sum = sum.Add(c.Price(b.price).Sub(avg).Pow(decimal.New(2, 0)))
	}
	return ebMath.Sqrt(sum.Div(decimal.New(int64(b.period), 0))), nil
}
