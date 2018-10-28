package macd

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/ema_mock"
	"eonbot/pkg/strategy/indicators/ma/ema"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMACDCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		MACD   *macd
		Result int
	}{
		{
			Name: "Successful count return",
			MACD: func() *macd {
				val, _ := New(3, 4, 2, 2, exchange.ClosePrice)
				return val
			}(),
			Result: 14,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.MACD.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestMACDCalc(t *testing.T) {
	tests := []struct {
		Name        string
		MACD        *macd
		Candles     []exchange.Candle
		Result      MACDInfo
		ShouldError bool
	}{
		{
			Name: "Unsuccessful MACD calculation of 12 candles when only 2 provided",
			MACD: func() *macd {
				val, _ := NewFromConfig(MACDConfig{EMA1Period: 3, EMA2Period: 4, SignalPeriod: 2, Price: exchange.ClosePrice}, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful MACD calculation of 12 candles when nil provided",
			MACD: func() *macd {
				val, _ := New(3, 4, 2, 0, exchange.ClosePrice)
				return val
			}(),
			Candles:     nil,
			ShouldError: true,
		},
		{
			Name: "Successful calculation of 12 candles",
			MACD: func() *macd {
				val, _ := New(3, 4, 2, 0, exchange.ClosePrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")},
				{Close: decimal.RequireFromString("13")},
				{Close: decimal.RequireFromString("14")},
				{Close: decimal.RequireFromString("15")},
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{Close: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{Close: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{Close: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
			},
			Result: MACDInfo{
				MACDLine:   decimal.RequireFromString("0.1112"),
				SignalLine: decimal.RequireFromString("0.1210"),
				Histogram:  decimal.RequireFromString("-0.0098"),
			},
		},
		{
			Name: "Successful calculation of 12 candles with offset set to 1",
			MACD: func() *macd {
				val, _ := New(3, 4, 2, 1, exchange.LowPrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.RequireFromString("10")},
				{Low: decimal.RequireFromString("11")},
				{Low: decimal.RequireFromString("12")},
				{Low: decimal.RequireFromString("13")},
				{Low: decimal.RequireFromString("14")},
				{Low: decimal.RequireFromString("15")},
				{Low: decimal.RequireFromString("10")},
				{Low: decimal.RequireFromString("11")},
				{Low: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{Low: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{Low: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{Low: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
				{Low: decimal.RequireFromString("11")},
			},
			Result: MACDInfo{
				MACDLine:   decimal.RequireFromString("0.1112"),
				SignalLine: decimal.RequireFromString("0.1210"),
				Histogram:  decimal.RequireFromString("-0.0098"),
			},
		},
		{
			Name: "Successful calculation of 12 candles with EMA2 being the fast one",
			MACD: func() *macd {
				val, _ := New(4, 3, 2, 0, exchange.HighPrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{High: decimal.RequireFromString("10")},
				{High: decimal.RequireFromString("11")},
				{High: decimal.RequireFromString("12")},
				{High: decimal.RequireFromString("13")},
				{High: decimal.RequireFromString("14")},
				{High: decimal.RequireFromString("15")},
				{High: decimal.RequireFromString("10")},
				{High: decimal.RequireFromString("11")},
				{High: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{High: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{High: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{High: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
			},
			Result: MACDInfo{
				MACDLine:   decimal.RequireFromString("0.1112"),
				SignalLine: decimal.RequireFromString("0.1210"),
				Histogram:  decimal.RequireFromString("-0.0098"),
			},
		},
		{
			Name: "Successful calculation of 12 candles with mocked EMA1 that returns error",
			MACD: func() *macd {
				ema2, _ := ema.New(3, 0, exchange.ClosePrice)
				sigEMA, _ := ema.New(2, 0, exchange.ClosePrice)
				val := newMACD(ema_mock.NewEMAMock(decimal.Zero, errors.New("test")), ema2, sigEMA, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")},
				{Close: decimal.RequireFromString("13")},
				{Close: decimal.RequireFromString("14")},
				{Close: decimal.RequireFromString("15")},
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{Close: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{Close: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{Close: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
			},
			ShouldError: true,
		},
		{
			Name: "Successful calculation of 12 candles with mocked EMA2 that returns error",
			MACD: func() *macd {
				ema1, _ := ema.New(4, 0, exchange.ClosePrice)
				sigEMA, _ := ema.New(2, 0, exchange.ClosePrice)
				val := newMACD(ema1, ema_mock.NewEMAMock(decimal.Zero, errors.New("test")), sigEMA, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")},
				{Close: decimal.RequireFromString("13")},
				{Close: decimal.RequireFromString("14")},
				{Close: decimal.RequireFromString("15")},
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{Close: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{Close: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{Close: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
			},
			ShouldError: true,
		},
		{
			Name: "Successful calculation of 12 candles with mocked SignalEMA that returns error",
			MACD: func() *macd {
				ema1, _ := ema.New(4, 0, exchange.ClosePrice)
				ema2, _ := ema.New(3, 0, exchange.ClosePrice)
				val := newMACD(ema1, ema2, ema_mock.NewEMAMockDec(decimal.Zero, errors.New("test")), 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")},
				{Close: decimal.RequireFromString("13")},
				{Close: decimal.RequireFromString("14")},
				{Close: decimal.RequireFromString("15")},
				{Close: decimal.RequireFromString("10")},
				{Close: decimal.RequireFromString("11")},
				{Close: decimal.RequireFromString("12")}, // 11.75 (EMA1) - 11.796 (EMA2) = -0.046
				{Close: decimal.RequireFromString("13")}, // 12.5 (EMA1) - 12.2776 (EMA2) = 0.2224
				{Close: decimal.RequireFromString("14")}, // 13.25 (EMA1) - 13.0832 (EMA2) = 0.1668
				{Close: decimal.RequireFromString("15")}, // 14 (EMA1) - 13.8888 (EMA2) = 0.1112
			},
			ShouldError: true,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.MACD.Calc(v.Candles)
			if !v.Result.MACDLine.Equal(res.MACDLine.Round(4)) {
				t.Errorf("incorrect MACDLine; \nexpected: %s, \ngot: %s", v.Result.MACDLine.String(), res.MACDLine.String())
			} else if !v.Result.SignalLine.Equal(res.SignalLine.Round(4)) {
				t.Errorf("incorrect signal line; \nexpected: %s, \ngot: %s", v.Result.SignalLine.String(), res.SignalLine.String())
			} else if !v.Result.Histogram.Equal(res.Histogram.Round(4)) {
				t.Errorf("incorrect histogram; \nexpected: %s, \ngot: %s", v.Result.Histogram.String(), res.Histogram.String())
			}
			if v.ShouldError {
				if err == nil {
					t.Error("error expected, but not returned")
				}
			} else {
				if err != nil {
					t.Error("error not expected, but returned:", err)
				}
			}
		})
	}
}
