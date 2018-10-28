package bb

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/sma_mock"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestBBNew(t *testing.T) {
	tests := []struct {
		Name        string
		MAType      string
		Price       string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful BB creation when MA type is invalid",
			MAType:      "test",
			Price:       exchange.ClosePrice,
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful BB creation when price type is invalid",
			MAType:      ma.SMAName,
			Price:       "test",
			ShouldError: true,
		},
		{
			Name:        "Successful BB creation",
			MAType:      ma.SMAName,
			Price:       exchange.ClosePrice,
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			_, err := New(3, 0, decimal.Zero, v.MAType, v.Price)
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

func TestBBCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		BB     *bb
		Result int
	}{
		{
			Name: "Successfully returned BB candles count",
			BB: func() *bb {
				newBB, _ := NewFromConfig(BBConfig{Period: 3, STDEV: decimal.Zero, Price: exchange.ClosePrice, MAType: ma.SMAName}, 2)
				return newBB
			}(),
			Result: 5,
		},
		{
			Name: "Successful count return",
			BB: func() *bb {
				val, _ := New(3, 2, decimal.Zero, ma.EMAName, exchange.ClosePrice)
				return val
			}(),
			Result: 8,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.BB.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestBBCalc(t *testing.T) {
	tests := []struct {
		Name        string
		BB          *bb
		Candles     []exchange.Candle
		Result      BBInfo
		ShouldError bool
	}{
		{
			Name: "Unsuccessful BB calc when no candles (nil) provided",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.ClosePrice)
				return newBB
			}(),
			Candles:     nil,
			Result:      BBInfo{decimal.Zero, decimal.Zero, decimal.Zero},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful BB calc of 3 candles when only 2 provided",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.ClosePrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(1, 0)},
				{Close: decimal.New(2, 0)},
			},
			Result:      BBInfo{decimal.Zero, decimal.Zero, decimal.Zero},
			ShouldError: true,
		},
		{
			Name: "Successful BB calc of 3 candles",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.LowPrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.New(1, 0)},
				{Low: decimal.New(2, 0)},
				{Low: decimal.New(3, 0)},
			},
			Result: BBInfo{
				decimal.RequireFromString("3.633"),
				decimal.RequireFromString("2"),
				decimal.RequireFromString("0.367"),
			},
		},
		{
			Name: "Successful BB calc of 3 candles with offset set to 1",
			BB: func() *bb {
				newBB, _ := New(3, 1, decimal.New(2, 0), ma.SMAName, exchange.LowPrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{Low: decimal.New(1, 0)},
				{Low: decimal.New(2, 0)},
				{Low: decimal.New(3, 0)},
				{Low: decimal.New(4, 0)},
			},
			Result: BBInfo{
				decimal.RequireFromString("3.633"),
				decimal.RequireFromString("2"),
				decimal.RequireFromString("0.367"),
			},
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.BB.Calc(v.Candles)
			if !v.Result.Upper.Equal(res.Upper.Round(4)) {
				t.Errorf("incorrect Upper line; \nexpected: %s, \ngot: %s", v.Result.Upper.String(), res.Upper.Round(4).String())
			} else if !v.Result.Middle.Equal(res.Middle.Round(4)) {
				t.Errorf("incorrect Middle line; \nexpected: %s, \ngot: %s", v.Result.Middle.String(), res.Middle.Round(4).String())
			} else if !v.Result.Lower.Equal(res.Lower.Round(4)) {
				t.Errorf("incorrect Lower line; \nexpected: %s, \ngot: %s", v.Result.Lower.String(), res.Lower.Round(4).String())
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

func TestBBStandardDeviation(t *testing.T) {
	tests := []struct {
		Name        string
		BB          *bb
		Candles     []exchange.Candle
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name: "Unsuccessful STDEV calc when no candles (nil) provided",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.ClosePrice)
				return newBB
			}(),
			Candles:     nil,
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful STDEV calc of 3 candles when only 2 provided",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.ClosePrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{Close: decimal.New(1, 0)},
				{Close: decimal.New(2, 0)},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful STDEV calc of 3 candles when SMA returns error",
			BB: newBB(3, 0, decimal.New(2, 0), exchange.HighPrice,
				sma_mock.NewSMAMock(decimal.Zero, nil, 3, 0),
				sma_mock.NewSMAMock(decimal.Zero, errors.New("test"), 3, 0)),
			Candles: []exchange.Candle{
				{High: decimal.New(1, 0)},
				{High: decimal.New(2, 0)},
				{High: decimal.New(3, 0)},
			},
			ShouldError: true,
		},
		{
			Name: "Successful STDEV calc of 3 candles",
			BB: func() *bb {
				newBB, _ := New(3, 0, decimal.New(2, 0), ma.SMAName, exchange.HighPrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{High: decimal.New(1, 0)},
				{High: decimal.New(2, 0)},
				{High: decimal.New(3, 0)},
			},
			Result: decimal.RequireFromString("0.8165"),
		},
		{
			Name: "Successful STDEV calc of 3 candles with offset set to 1",
			BB: func() *bb {
				newBB, _ := New(3, 1, decimal.New(2, 0), ma.SMAName, exchange.HighPrice)
				return newBB
			}(),
			Candles: []exchange.Candle{
				{High: decimal.New(1, 0)},
				{High: decimal.New(2, 0)},
				{High: decimal.New(3, 0)},
				{High: decimal.New(4, 0)},
			},
			Result: decimal.RequireFromString("0.8165"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.BB.standardDeviation(v.Candles)
			if !v.Result.Equal(res.Round(4)) {
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
