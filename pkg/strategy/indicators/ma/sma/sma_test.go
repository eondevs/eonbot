package sma

import (
	"eonbot/pkg/exchange"
	"testing"

	"github.com/shopspring/decimal"
)

func TestSMACandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		SMA    *sma
		Result int
	}{
		{
			Name: "Successful count return",
			SMA: func() *sma {
				val, _ := New(3, 2, exchange.ClosePrice)
				return val
			}(),
			Result: 5,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.SMA.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestSMANew(t *testing.T) {
	tests := []struct {
		Name        string
		Period      int
		Offset      int
		Price       string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful SMA creation when price type is invalid",
			Period:      1,
			Offset:      1,
			Price:       "test",
			ShouldError: true,
		},
		{
			Name:        "Successful SMA creation",
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

func TestSMACalc(t *testing.T) {
	tests := []struct {
		Name        string
		SMA         *sma
		Candles     []exchange.Candle
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful SMA calc of 6 candles when only 4 are provided",
			SMA: func() *sma {
				ma, _ := New(6, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(20, 0)},
				{Close: decimal.New(30, 0)},
				{Close: decimal.New(40, 0)},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful SMA calc of 3 candles when nil provided",
			SMA: func() *sma {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful SMA calc of 3 candles",
			SMA: func() *sma {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(20, 0)},
				{Close: decimal.New(30, 0)},
			},
			Result: decimal.New(20, 0),
		},
		{
			Name: "Successful SMA calc of 6 candles",
			SMA: func() *sma {
				ma, _ := New(6, 0, exchange.OpenPrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Open: decimal.New(10, 0)},
				{Open: decimal.New(20, 0)},
				{Open: decimal.New(30, 0)},
				{Open: decimal.New(40, 0)},
				{Open: decimal.New(50, 0)},
				{Open: decimal.New(60, 0)},
			},
			Result: decimal.New(35, 0),
		},
		{
			Name: "Successful SMA calc of 3 candles with offset set to 2",
			SMA: func() *sma {
				ma, _ := New(3, 2, exchange.HighPrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{High: decimal.New(10, 0)},
				{High: decimal.New(20, 0)},
				{High: decimal.New(30, 0)},
				{High: decimal.New(40, 0)},
				{High: decimal.New(50, 0)},
				{High: decimal.New(60, 0)},
			},
			Result: decimal.New(30, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.SMA.Calc(v.Candles)
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

func TestSMACalcDecimal(t *testing.T) {
	tests := []struct {
		Name        string
		SMA         *sma
		Values      []decimal.Decimal
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful SMA calc of 6 values when only 4 are provided",
			SMA: func() *sma {
				ma, _ := New(6, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
				decimal.New(40, 0),
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful SMA calc of 6 values when nil provided",
			SMA: func() *sma {
				ma, _ := New(6, 0, exchange.ClosePrice)
				return ma
			}(),
			Values:      nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful SMA calc of 3 values",
			SMA: func() *sma {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
			},
			Result: decimal.New(20, 0),
		},
		{
			Name: "Successful SMA calc of 6 values",
			SMA: func() *sma {
				ma, _ := New(6, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
				decimal.New(40, 0),
				decimal.New(50, 0),
				decimal.New(60, 0),
			},
			Result: decimal.New(35, 0),
		},
		{
			Name: "Successful SMA calc of 3 values with offset set to 2",
			SMA: func() *sma {
				ma, _ := New(3, 2, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(30, 0),
				decimal.New(40, 0),
				decimal.New(50, 0),
				decimal.New(60, 0),
			},
			Result: decimal.New(30, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.SMA.CalcDecimal(v.Values)
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
