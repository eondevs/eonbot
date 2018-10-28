package rsi

import (
	"eonbot/pkg/exchange"
	"testing"

	"github.com/shopspring/decimal"
)

func TestRSICandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		RSI    *rsi
		Result int
	}{
		{
			Name: "Successful count return",
			RSI: func() *rsi {
				val, _ := New(3, 2, exchange.ClosePrice)
				return val
			}(),
			Result: 8,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.RSI.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestRSINew(t *testing.T) {
	tests := []struct {
		Name        string
		Period      int
		Offset      int
		Price       string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful RSI creation when price type is invalid",
			Period:      1,
			Offset:      1,
			Price:       "test",
			ShouldError: true,
		},
		{
			Name:        "Successful RSI creation",
			Period:      1,
			Offset:      1,
			Price:       exchange.ClosePrice,
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			_, err := New(v.Period, v.Offset, v.Price)
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

func TestRSICalc(t *testing.T) {
	tests := []struct {
		Name        string
		RSI         *rsi
		Candles     []exchange.Candle
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful calculation of 5 candles RSI when only 2 provided",
			RSI: func() *rsi {
				val, _ := NewFromConfig(RSIConfig{Period: 5, Price: exchange.ClosePrice}, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(4, 0)},
				{Close: decimal.New(3, 0)},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful calculation of 5 candles RSI when nil provided",
			RSI: func() *rsi {
				val, _ := New(5, 0, exchange.ClosePrice)
				return val
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successfully calculated 2 candles RSI",
			RSI: func() *rsi {
				val, _ := New(2, 0, exchange.ClosePrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(4, 0)},
				{Close: decimal.New(3, 0)},
				{Close: decimal.New(5, 0)},
				{Close: decimal.New(6, 0)},
			},
			Result: decimal.RequireFromString("80"),
		},
		{
			Name: "Successfully calculated 2 candles RSI with offset set to 2",
			RSI: func() *rsi {
				val, _ := New(2, 2, exchange.HighPrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{High: decimal.New(4, 0)},
				{High: decimal.New(3, 0)},
				{High: decimal.New(5, 0)},
				{High: decimal.New(6, 0)},
				{High: decimal.New(8, 0)},
				{High: decimal.New(3, 0)},
			},
			Result: decimal.RequireFromString("80"),
		},
		{
			Name: "Successfully calculated 14 candles RSI",
			RSI: func() *rsi {
				val, _ := New(14, 0, exchange.LowPrice)
				return val
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.RequireFromString("44.34")},
				{Low: decimal.RequireFromString("44.09")},
				{Low: decimal.RequireFromString("44.15")},
				{Low: decimal.RequireFromString("43.61")},
				{Low: decimal.RequireFromString("44.33")},
				{Low: decimal.RequireFromString("44.83")},
				{Low: decimal.RequireFromString("45.10")},
				{Low: decimal.RequireFromString("45.42")},
				{Low: decimal.RequireFromString("45.84")},
				{Low: decimal.RequireFromString("46.08")},
				{Low: decimal.RequireFromString("45.89")},
				{Low: decimal.RequireFromString("46.03")},
				{Low: decimal.RequireFromString("45.61")},
				{Low: decimal.RequireFromString("46.28")},
				{Low: decimal.RequireFromString("46.28")},
				{Low: decimal.RequireFromString("46")},
				{Low: decimal.RequireFromString("46.03")},
				{Low: decimal.RequireFromString("46.41")},
				{Low: decimal.RequireFromString("46.22")},
				{Low: decimal.RequireFromString("45.64")},
				{Low: decimal.RequireFromString("46.21")},
				{Low: decimal.RequireFromString("46.25")},
				{Low: decimal.RequireFromString("45.71")},
				{Low: decimal.RequireFromString("46.45")},
				{Low: decimal.RequireFromString("45.78")},
				{Low: decimal.RequireFromString("45.35")},
				{Low: decimal.RequireFromString("44.03")},
				{Low: decimal.RequireFromString("44.18")},
			},
			Result: decimal.RequireFromString("41.49"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.RSI.Calc(v.Candles)
			if !v.Result.Equal(res.Round(2)) {
				t.Errorf("incorrect result; expected: %s, got: %s", v.Result.String(), res.String())
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
