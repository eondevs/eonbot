package ema_mock

import (
	"eonbot/pkg/exchange"

	"github.com/shopspring/decimal"
)

type emaMock struct {
	err    error
	val    decimal.Decimal
	errDec error
	valDec decimal.Decimal
}

func NewEMAMock(val decimal.Decimal, err error) *emaMock {
	return &emaMock{val: val, err: err}
}

func NewEMAMockDec(val decimal.Decimal, err error) *emaMock {
	return &emaMock{valDec: val, errDec: err}
}

func (e *emaMock) CandlesCount() int {
	return 0
}

func (e *emaMock) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	if e.err != nil {
		return decimal.Zero, e.err
	}
	return e.val, nil
}

func (e *emaMock) CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error) {
	if e.errDec != nil {
		return decimal.Zero, e.errDec
	}
	return e.valDec, nil
}
