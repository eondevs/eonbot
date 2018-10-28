package rsi_mock

import (
	"eonbot/pkg/exchange"

	"github.com/shopspring/decimal"
)

type rsiMock struct {
	err    error
	val    decimal.Decimal
	period int
	offset int
}

func NewRSIMock(val decimal.Decimal, err error, period, offset int) *rsiMock {
	return &rsiMock{val: val, err: err, period: period, offset: offset}
}

func (r *rsiMock) CandlesCount() int {
	return r.period*2 + r.offset
}

func (r *rsiMock) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	if r.err != nil {
		return decimal.Zero, r.err
	}
	return r.val, nil
}
