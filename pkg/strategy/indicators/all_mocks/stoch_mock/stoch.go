package stoch_mock

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/stoch"
)

type stochMock struct {
	err     error
	val     stoch.StochInfo
	kPeriod int
	dPeriod int
	offset  int
}

func NewStochMock(val stoch.StochInfo, err error, kPeriod, dPeriod, offset int) *stochMock {
	return &stochMock{val: val, err: err, kPeriod: kPeriod, dPeriod: dPeriod, offset: offset}
}

func (s *stochMock) CandlesCount() int {
	return s.kPeriod + (s.dPeriod - 1) + s.offset
}

func (s *stochMock) Calc(cc []exchange.Candle) (stoch.StochInfo, error) {
	if s.err != nil {
		return stoch.StochInfo{}, s.err
	}
	return s.val, nil
}
