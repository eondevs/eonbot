package wma

import (
	"eonbot/pkg/exchange"
	"testing"

	"github.com/shopspring/decimal"
)

func TestWMACandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		WMA    *wma
		Result int
	}{
		{
			Name: "Successful count return",
			WMA: func() *wma {
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
			res := v.WMA.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestWMANew(t *testing.T) {
	tests := []struct {
		Name        string
		Period      int
		Offset      int
		Price       string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful WMA creation when price type is invalid",
			Period:      1,
			Offset:      1,
			Price:       "test",
			ShouldError: true,
		},
		{
			Name:        "Successful WMA creation",
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

func TestWMACalc(t *testing.T) {
	tests := []struct {
		Name        string
		WMA         *wma
		Candles     []exchange.Candle
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful WMA calc of 4 candles when only 2 candles provided",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.ClosePrice)
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
			Name: "Unsuccessful WMA calc of 4 candles when nil provided",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful WMA calc of 3 candles",
			WMA: func() *wma {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(10, 0)},
				{Close: decimal.New(40, 0)},
				{Close: decimal.New(30, 0)},
			},
			Result: decimal.New(30, 0),
		},
		{
			Name: "Successful WMA calc of 3 candles with offset set to 1",
			WMA: func() *wma {
				ma, _ := New(3, 1, exchange.OpenPrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Open: decimal.New(10, 0)},
				{Open: decimal.New(40, 0)},
				{Open: decimal.New(30, 0)},
				{Open: decimal.New(5, 0)},
			},
			Result: decimal.New(30, 0),
		},
		{
			Name: "Successful WMA calc of 4 candles",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.LowPrice)
				return ma
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.New(10, 0)},
				{Low: decimal.New(20, 0)},
				{Low: decimal.New(10, 0)},
				{Low: decimal.New(5, 0)},
			},
			Result: decimal.New(10, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.WMA.Calc(v.Candles)
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

func TestWMACalcDecimal(t *testing.T) {
	tests := []struct {
		Name        string
		WMA         *wma
		Values      []decimal.Decimal
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful WMA calc of 4 values when only 2 values provided",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.ClosePrice)
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
			Name: "Unsuccessful WMA calc of 4 values when nil provided",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.ClosePrice)
				return ma
			}(),
			Values:      nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful WMA calc of 3 values",
			WMA: func() *wma {
				ma, _ := New(3, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(40, 0),
				decimal.New(30, 0),
			},
			Result: decimal.New(30, 0),
		},
		{
			Name: "Successful WMA calc of 3 values with offset set to 1",
			WMA: func() *wma {
				ma, _ := New(3, 1, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(40, 0),
				decimal.New(30, 0),
				decimal.New(5, 0),
			},
			Result: decimal.New(30, 0),
		},
		{
			Name: "Successful WMA calc of 4 values",
			WMA: func() *wma {
				ma, _ := New(4, 0, exchange.ClosePrice)
				return ma
			}(),
			Values: []decimal.Decimal{
				decimal.New(10, 0),
				decimal.New(20, 0),
				decimal.New(10, 0),
				decimal.New(5, 0),
			},
			Result: decimal.New(10, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.WMA.CalcDecimal(v.Values)
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
