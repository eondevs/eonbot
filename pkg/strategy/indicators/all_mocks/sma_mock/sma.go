package sma_mock

import (
	"eonbot/pkg/exchange"

	"github.com/shopspring/decimal"
)

type smaMock struct {
	err    error
	val    decimal.Decimal
	errDec error
	valDec decimal.Decimal
	period int
	offset int
}

func NewSMAMock(val decimal.Decimal, err error, period, offset int) *smaMock {
	return &smaMock{val: val, err: err, period: period, offset: offset}
}

func NewSMAMockDec(val decimal.Decimal, err error, period, offset int) *smaMock {
	return &smaMock{valDec: val, errDec: err, period: period, offset: offset}
}

func (s *smaMock) CandlesCount() int {
	return s.period + s.offset
}

func (s *smaMock) Calc(cc []exchange.Candle) (decimal.Decimal, error) {
	if s.err != nil {
		return decimal.Zero, s.err
	}
	return s.val, nil
}

func (s *smaMock) CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error) {
	if s.errDec != nil {
		return decimal.Zero, s.errDec
	}
	return s.valDec, nil
}
