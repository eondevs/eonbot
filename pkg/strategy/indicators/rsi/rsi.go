// Package rsi implements RSI indicator calculation logic.
package rsi

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/utils"
	"errors"

	"github.com/shopspring/decimal"
)

type RSI interface {
	CandlesCount() int
	Calc(cc []exchange.Candle) (decimal.Decimal, error)
}

// RSI contains internal data values needed
// to calculate RSI.
type rsi struct {
	period int
	offset int
	price  string
}

// New creates new RSI object with provided period
// and offset to make further calculations.
func New(period, offset int, price string) (*rsi, error) {
	if err := exchange.CandlePriceValid(price); err != nil {
		return nil, err
	}

	return &rsi{
		period: period,
		offset: offset,
		price:  price,
	}, nil
}

// NewFromConfig creates new RSI object the same way as New,
// it just takes values from provided RSIConfig.
func NewFromConfig(conf RSIConfig, offset int) (*rsi, error) {
	return New(conf.Period, offset, conf.Price)
}

// CandlesCount returns min candle count needed
// to calculate RSI with the provided period.
func (r *rsi) CandlesCount() int {
	return r.period*2 + r.offset
}

// Calc calculates RSI of provided period values.
// It will slice out only needed candles (period and offset
// are used to calc boundaries).
// Returns RSI calculation result and optionally an error.
// RSI calculation:
// 0. Gain/Loss is calculated by subtracting previous candle price
//      from current candle's price. If the result is negative -
//      it's a loss (when average loss is calculated, absolute value of
//      the result is used), if positive - it's a gain;
// 1. Average Gain/Loss calc (X in this example represents period count):
//    * First Average Gain = Sum of Gains over the past X periods / X;
//    * First Average Loss = Sum of Losses over the past X periods / X;
//    -----
//    * Average Gain = ((previous Average Gain) x (X - 1) + current Gain) / X;
//    * Average Loss = ((previous Average Loss) x (X - 1) + current Loss) / X;
// 2. RS = AverageGain/AverageLoss;
// 3. RSI = 100 - 100 / (1 - RS);
func (r *rsi) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	start := r.CandlesCount()
	end := r.offset

	if cc == nil || len(cc) < start {
		return decimal.Zero, errors.New("RSI candles list is too small")
	}

	candles := cc[len(cc)-start : len(cc)-end]

	prevGain := decimal.Zero
	prevLoss := decimal.Zero

	// First N candles' averages cannot be calculated (too few candles before them).
	// We start from r.period and not r.period-1 because gain/loss is calculated by
	// using previous candle value and current candle value.
	for i := r.period; i < len(candles); i++ {
		currentGain := decimal.Zero
		currentLoss := decimal.Zero

		calc := func(val1, val2 decimal.Decimal) {
			change := val1.Sub(val2)
			if change.Sign() > 0 {
				currentGain = currentGain.Add(change)
			} else if change.Sign() < 0 {
				currentLoss = currentLoss.Add(change.Abs())
			}
		}

		// first candle of RSI period (since it's the first period entry
		// it won't have gain/loss, but it will have price which will be
		// used in the second entry to calc gain/loss).
		if i == r.period {
			// use first N candles to calc first RSI candle gain/loss
			for j := i - r.period + 1; j <= i; j++ {
				calc(candles[j].Price(r.price), candles[j-1].Price(r.price))
			}
			prevGain = currentGain.Div(decimal.New(int64(r.period), 0))
			prevLoss = currentLoss.Div(decimal.New(int64(r.period), 0))
			continue
		}
		calc(candles[i].Price(r.price), candles[i-1].Price(r.price))

		prevGain = prevGain.Mul(decimal.New(int64(r.period-1), 0)).Add(currentGain).Div(decimal.New(int64(r.period), 0))
		prevLoss = prevLoss.Mul(decimal.New(int64(r.period-1), 0)).Add(currentLoss).Div(decimal.New(int64(r.period), 0))
	}

	rs := prevGain.Div(utils.PreventZero(prevLoss))
	return decimal.New(100, 0).Sub(decimal.New(100, 0).Div(utils.PreventZero(decimal.New(1, 0).Add(rs)))), nil
}
