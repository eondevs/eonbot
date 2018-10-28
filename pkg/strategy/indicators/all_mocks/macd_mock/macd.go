package macd_mock

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/macd"
)

type macdMock struct {
	err    error
	val    macd.MACDInfo
	ema1   int
	ema2   int
	signal int
	offset int
}

func NewMACDMock(val macd.MACDInfo, err error, ema1, ema2, signal, offset int) *macdMock {
	return &macdMock{val: val, err: err, ema1: ema1, ema2: ema2, signal: signal, offset: offset}
}

func (m *macdMock) CandlesCount() (count int) {
	if m.ema1*2 > m.ema2*2 {
		count = m.ema1 * 2
	} else {
		count = m.ema2 * 2
	}
	count += m.signal*2 + m.offset
	return count
}

func (m *macdMock) Calc(cc []exchange.Candle) (macd.MACDInfo, error) {
	if m.err != nil {
		return macd.MACDInfo{}, m.err
	}
	return m.val, nil
}
