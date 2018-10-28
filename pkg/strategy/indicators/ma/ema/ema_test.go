package ema

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/sma_mock"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEMACandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		EMA    *ema
		Result int
	}{
		{
			Name: "Successful count return",
			EMA: func() *ema {
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
			res := v.EMA.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestEMANew(t *testing.T) {
	tests := []struct {
		Name        string
		Period      int
		Offset      int
		Price       string
		ShouldError bool
		PrivateNew  bool
	}{
		{
			Name:        "Unsuccessful EMA creation when price type is invalid",
			Period:      1,
			Offset:      1,
			Price:       "test",
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful EMA creation with private func when price type is invalid",
			Period:      1,
			Offset:      1,
			Price:       "test",
			ShouldError: true,
			PrivateNew:  true,
		},
		{
			Name:        "Successful EMA creation",
			Period:      1,
			Offset:      1,
			Price:       exchange.ClosePrice,
			ShouldError: false,
		},
		{
			Name:        "Successful EMA creation",
			Period:      1,
			Offset:      1,
			Price:       exchange.ClosePrice,
			ShouldError: false,
			PrivateNew:  true,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			var err error
			if v.PrivateNew {
				_, err = newEMA(v.Period, v.Offset, v.Price, sma_mock.NewSMAMock(decimal.Zero, nil, 2, 0))
			} else {
				_, err = New(v.Period, v.Offset, v.Price)
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

func TestEMACalc(t *testing.T) {
	tests := []struct {
		Name        string
		EMA         *ema
		Candles     []exchange.Candle
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful EMA calc of 3 candles when only 2 candles provided",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMock(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(20, 0)},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA calc of 3 candles when nil provided",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMock(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA calc of 3 candles when only 2 out of 6 candles provided",
			EMA: func() *ema {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(20, 0)},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA with real SMA calc of 3 candles when nil provided",
			EMA: func() *ema {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful EMA calc of 3 candles",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMock(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(20, 0)},
				{Close: decimal.New(30, 0)},
			},
			Result: decimal.RequireFromString("22.5"),
		},
		{
			Name: "Successful EMA calc of 3 candles with offset set to 1",
			EMA: func() *ema {
				ma, _ := newEMA(3, 1, exchange.OpenPrice, sma_mock.NewSMAMock(decimal.New(10, 0), nil, 3, 4))
				return ma
			}(),
			Candles: []exchange.Candle{
				{Open: decimal.New(10, 0)},
				{Open: decimal.New(20, 0)},
				{Open: decimal.New(30, 0)},
				{Open: decimal.New(40, 0)},
			},
			Result: decimal.RequireFromString("22.5"),
		},
		{
			Name: "Successful EMA with real SMA calc of 3 candles with offset set to 1",
			EMA: func() *ema {
				ma, _ := New(3, 1, exchange.HighPrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{High: decimal.New(10, 0)},
				{High: decimal.New(11, 0)},
				{High: decimal.New(12, 0)},
				{High: decimal.New(13, 0)},
				{High: decimal.New(14, 0)},
				{High: decimal.New(15, 0)},
				{High: decimal.New(16, 0)},
			},
			Result: decimal.RequireFromString("14"),
		},
		{
			Name: "Successful EMA calc of 4 candles",
			EMA: func() *ema {
				ma, _ := newEMA(4, 0, exchange.LowPrice, sma_mock.NewSMAMock(decimal.New(10, 0), nil, 4, 4))
				return ma
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.New(10, 0)},
				{Low: decimal.New(20, 0)},
				{Low: decimal.New(30, 0)},
				{Low: decimal.New(40, 0)},
			},
			Result: decimal.RequireFromString("28.24"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.EMA.Calc(v.Candles)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, \nexpected: %s, \ngot: %s", v.Result.String(), res.String())
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

func TestEMACalcDecimal(t *testing.T) {
	tests := []struct {
		Name        string
		EMA         *ema
		Values      []decimal.Decimal
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful EMA calc of 3 values when only 2 values provided",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMockDec(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA calc of 3 values when nil provided",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMockDec(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Values:      nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA calc of 3 values when only 2 out of 6 values provided",
			EMA: func() *ema {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful EMA with real SMA calc of 3 values when nil provided",
			EMA: func() *ema {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Values:      nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful EMA calc of 3 values",
			EMA: func() *ema {
				ma, _ := newEMA(3, 0, exchange.ClosePrice, sma_mock.NewSMAMockDec(decimal.New(10, 0), nil, 3, 3))
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
			},
			Result: decimal.RequireFromString("22.5"),
		},
		{
			Name: "Successful EMA calc of 3 values with offset set to 1",
			EMA: func() *ema {
				ma, _ := newEMA(3, 1, exchange.ClosePrice, sma_mock.NewSMAMockDec(decimal.New(10, 0), nil, 3, 4))
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
				decimal.New(40, 0),
			},
			Result: decimal.RequireFromString("22.5"),
		},
		{
			Name: "Successful EMA with real SMA calc of 3 values with offset set to 1",
			EMA: func() *ema {
				ma, _ := New(3, 1, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(11, 0),
				decimal.New(12, 0),
				decimal.New(13, 0),
				decimal.New(14, 0),
				decimal.New(15, 0),
				decimal.New(16, 0),
			},
			Result: decimal.RequireFromString("14"),
		},
		{
			Name: "Successful EMA calc of 4 values",
			EMA: func() *ema {
				ma, _ := newEMA(4, 0, exchange.ClosePrice, sma_mock.NewSMAMockDec(decimal.New(10, 0), nil, 4, 4))
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
				decimal.New(40, 0),
			},
			Result: decimal.RequireFromString("28.24"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.EMA.CalcDecimal(v.Values)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, \nexpected: %s, \ngot: %s", v.Result.String(), res.String())
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
