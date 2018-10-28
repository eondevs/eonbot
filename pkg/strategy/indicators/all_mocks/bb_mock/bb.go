package bb_mock

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/bb"
)

type bbMock struct {
	err    error
	val    bb.BBInfo
	period int
	offset int
}

func NewBBMock(val bb.BBInfo, err error, period, offset int) *bbMock {
	return &bbMock{val: val, err: err, period: period, offset: offset}
}

func (b *bbMock) CandlesCount() int {
	return b.period + b.offset // SMA is 'used' for this mock
}

func (b *bbMock) Calc(cc []exchange.Candle) (bb.BBInfo, error) {
	if b.err != nil {
		return bb.BBInfo{}, b.err
	}
	return b.val, nil
}
