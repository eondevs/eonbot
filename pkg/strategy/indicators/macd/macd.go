// Package macd implements MACD indicator calculation logic.
package macd

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma/ema"
	"errors"

	"github.com/shopspring/decimal"
)

type MACD interface {
	CandlesCount() int
	Calc(cc []exchange.Candle) (MACDInfo, error)
}

// MACD contains ema1, ema2, signal EMAs and
// offset values needed to calculate MACDInfo.
type macd struct {
	offset    int
	ema1      ema.EMA
	ema2      ema.EMA
	signalEMA ema.EMA
}

// MACDInfo contains data calculated
// by Calc function.
type MACDInfo struct {
	MACDLine   decimal.Decimal `json:"macdLine"`
	SignalLine decimal.Decimal `json:"signalLine"`
	Histogram  decimal.Decimal `json:"-"`
}

// New creates new MACD object with provided period
// and offset to make further calculations.
func New(ema1, ema2, signal, offset int, price string) (*macd, error) {
	// NOTE: offset is used in Calc, so it's not needed
	// during EMAs creation.

	ema1Indicator, err := ema.New(ema1, 0, price)
	if err != nil {
		return nil, err
	}

	ema2Indicator, err := ema.New(ema2, 0, price)
	if err != nil {
		return nil, err
	}

	signalEMA, err := ema.New(signal, 0, price)
	if err != nil {
		return nil, err
	}

	return newMACD(
		ema1Indicator,
		ema2Indicator,
		signalEMA,
		offset), nil
}

// NewFromConfig creates new MACD object the same way as New,
// it just takes values from provided MACDConfig.
func NewFromConfig(conf MACDConfig, offset int) (*macd, error) {
	return New(conf.EMA1Period, conf.EMA2Period, conf.SignalPeriod, offset, conf.Price)
}

// newMACD creates new MACD object with specfied data values.
func newMACD(ema1, ema2, signalEMA ema.EMA, offset int) *macd {
	return &macd{
		ema1:      ema1,
		ema2:      ema2,
		signalEMA: signalEMA,
		offset:    offset,
	}
}

// CandlesCount returns min candle count needed
// to calculate MACD with the specified EMA and signal periods.
func (m *macd) CandlesCount() (count int) {
	if m.ema1.CandlesCount() > m.ema2.CandlesCount() {
		count = m.ema1.CandlesCount()
	} else {
		count = m.ema2.CandlesCount()
	}
	count += m.signalEMA.CandlesCount() + m.offset
	return count
}

// CalcMACD calculates MACD data(MACD line, signal line and histogram)
// of provided period candles. It will slice out only needed candles
// (period and offset are used to calc boundaries).
// Returns MACDInfo calculation result and optionally an error.
// MACD calculation:
// 1. MACD line = fast EMA (e.g. 12 candles) - slow EMA (e.g. 26 candles);
// 2. Signal line = EMA (e.g. 9 values) of MACD line (e.g. 9 MACD values from
//      calculations above, though since it's an EMA and first element's EMA
//      is SMA, 2 * signalPeriod values will be required, which in this
//      example would be 18);
// 3.Histogram = MACD line - Signal line;
func (m *macd) Calc(cc []exchange.Candle) (MACDInfo, error) {
	start := m.CandlesCount()
	end := m.offset

	if cc == nil || len(cc) < start {
		return MACDInfo{}, errors.New("MACD candles list is too small")
	}

	candles := cc[len(cc)-start : len(cc)-end]

	var macd []decimal.Decimal
	sigInit := m.signalEMA.CandlesCount() - 1
	for i := 0; i <= sigInit; i++ {
		ema1Res, err := m.ema1.Calc(candles[i : len(candles)-(sigInit-i)])
		if err != nil {
			return MACDInfo{}, err
		}

		ema2Res, err := m.ema2.Calc(candles[i : len(candles)-(sigInit-i)])
		if err != nil {
			return MACDInfo{}, err
		}

		macdVal := decimal.Zero
		if m.ema1.CandlesCount() > m.ema2.CandlesCount() {
			macdVal = ema2Res.Sub(ema1Res)
		} else {
			macdVal = ema1Res.Sub(ema2Res)
		}

		macd = append(macd, macdVal)
	}

	lastSignal, err := m.signalEMA.CalcDecimal(macd)
	if err != nil {
		return MACDInfo{}, err
	}

	lastMACD := macd[len(macd)-1] // CalcFullEMAFromDecimal checks if slice is initialized

	return MACDInfo{
		MACDLine:   lastMACD,
		SignalLine: lastSignal,
		Histogram:  lastMACD.Sub(lastSignal),
	}, nil
}
