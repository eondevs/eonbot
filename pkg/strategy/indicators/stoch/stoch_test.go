package stoch

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/sma_mock"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestStochCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		Stoch  *stoch
		Result int
	}{
		{
			Name: "Successful count return",
			Stoch: func() *stoch {
				val, _ := New(14, 3, 2)
				return val
			}(),
			Result: 18,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.Stoch.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestStochCalc(t *testing.T) {
	tests := []struct {
		Name        string
		Stoch       *stoch
		Candles     []exchange.Candle
		Result      StochInfo
		ShouldError bool
	}{
		{
			Name: "Unsuccessful calculation of 7 candles Stoch when only 2 provided",
			Stoch: func() *stoch {
				val, _ := NewFromConfig(StochConfig{KPeriod: 5, DPeriod: 3}, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{
					High:  decimal.RequireFromString("11"),
					Low:   decimal.RequireFromString("12"),
					Close: decimal.RequireFromString("10"),
				},
				{
					High:  decimal.RequireFromString("9"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful calculation of 7 candles Stoch when nil provided",
			Stoch: func() *stoch {
				val, _ := New(5, 3, 0)
				return val
			}(),
			Candles:     nil,
			ShouldError: true,
		},
		{
			Name:  "Unsuccessful calculation of 4 candles Stoch when SMA calc returns error",
			Stoch: newStoch(2, 3, 0, sma_mock.NewSMAMockDec(decimal.Zero, errors.New("test"), 3, 0)),
			Candles: []exchange.Candle{
				{
					High:  decimal.RequireFromString("12"),
					Low:   decimal.RequireFromString("11"),
					Close: decimal.RequireFromString("9"),
				},
				{
					High:  decimal.RequireFromString("9"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("10"),
				},
				{
					High:  decimal.RequireFromString("13"),
					Low:   decimal.RequireFromString("7"),
					Close: decimal.RequireFromString("12"),
				},
				{
					High:  decimal.RequireFromString("10"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("9"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Successful calculation of 4 candles Stoch",
			Stoch: func() *stoch {
				val, _ := New(2, 3, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{
					High:  decimal.RequireFromString("12"),
					Low:   decimal.RequireFromString("11"),
					Close: decimal.RequireFromString("9"),
				},
				{
					High:  decimal.RequireFromString("9"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("10"),
				},
				{
					High:  decimal.RequireFromString("13"),
					Low:   decimal.RequireFromString("7"),
					Close: decimal.RequireFromString("12"),
				},
				{
					High:  decimal.RequireFromString("10"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("9"),
				},
			},
			Result: StochInfo{
				K: decimal.RequireFromString("33.333"),
				D: decimal.RequireFromString("55.556"),
			},
		},
		{
			Name: "Successful calculation of 4 candles Stoch when offset set to 2",
			Stoch: func() *stoch {
				val, _ := New(2, 3, 2)
				return val
			}(),
			Candles: []exchange.Candle{
				{
					High:  decimal.RequireFromString("12"),
					Low:   decimal.RequireFromString("11"),
					Close: decimal.RequireFromString("9"),
				},
				{
					High:  decimal.RequireFromString("9"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("10"),
				},
				{
					High:  decimal.RequireFromString("13"),
					Low:   decimal.RequireFromString("7"),
					Close: decimal.RequireFromString("12"),
				},
				{
					High:  decimal.RequireFromString("10"),
					Low:   decimal.RequireFromString("8"),
					Close: decimal.RequireFromString("9"),
				},
				{
					High:  decimal.RequireFromString("100"),
					Low:   decimal.RequireFromString("70"),
					Close: decimal.RequireFromString("120"),
				},
				{
					High:  decimal.RequireFromString("103"),
					Low:   decimal.RequireFromString("80"),
					Close: decimal.RequireFromString("90"),
				},
			},
			Result: StochInfo{
				K: decimal.RequireFromString("33.333"),
				D: decimal.RequireFromString("55.556"),
			},
		},
		{
			Name: "Successful calculation of 14 candles Stoch",
			Stoch: func() *stoch {
				val, _ := New(14, 3, 0)
				return val
			}(),
			Candles: []exchange.Candle{
				{
					High: decimal.RequireFromString("127.01"),
					Low:  decimal.RequireFromString("125.36"),
				},
				{
					High: decimal.RequireFromString("127.62"),
					Low:  decimal.RequireFromString("126.16"),
				},
				{
					High: decimal.RequireFromString("126.59"),
					Low:  decimal.RequireFromString("124.93"),
				},
				{
					High: decimal.RequireFromString("127.35"),
					Low:  decimal.RequireFromString("126.09"),
				},
				{
					High: decimal.RequireFromString("128.17"),
					Low:  decimal.RequireFromString("126.82"),
				},
				{
					High: decimal.RequireFromString("128.43"),
					Low:  decimal.RequireFromString("126.48"),
				},
				{
					High: decimal.RequireFromString("127.37"),
					Low:  decimal.RequireFromString("126.03"),
				},
				{
					High: decimal.RequireFromString("126.42"),
					Low:  decimal.RequireFromString("124.83"),
				},
				{
					High: decimal.RequireFromString("126.90"),
					Low:  decimal.RequireFromString("126.39"),
				},
				{
					High: decimal.RequireFromString("126.85"),
					Low:  decimal.RequireFromString("125.72"),
				},
				{
					High: decimal.RequireFromString("125.65"),
					Low:  decimal.RequireFromString("124.56"),
				},
				{
					High: decimal.RequireFromString("125.72"),
					Low:  decimal.RequireFromString("124.57"),
				},
				{
					High: decimal.RequireFromString("127.16"),
					Low:  decimal.RequireFromString("125.07"),
				},
				{
					High:  decimal.RequireFromString("127.72"),
					Low:   decimal.RequireFromString("126.86"),
					Close: decimal.RequireFromString("127.29"),
				},
				{
					High:  decimal.RequireFromString("127.69"),
					Low:   decimal.RequireFromString("126.63"),
					Close: decimal.RequireFromString("127.18"),
				},
				{
					High:  decimal.RequireFromString("128.22"),
					Low:   decimal.RequireFromString("126.80"),
					Close: decimal.RequireFromString("128.01"),
				},
			},
			Result: StochInfo{
				K: decimal.RequireFromString("89.147"),
				D: decimal.RequireFromString("75.797"),
			},
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.Stoch.Calc(v.Candles)
			if !v.Result.K.Equal(res.K.Round(3)) {
				t.Errorf("incorrect K; expected: %s, got: %s", v.Result.K.String(), res.K.String())
			} else if !v.Result.D.Equal(res.D.Round(3)) {
				t.Errorf("incorrect D; expected: %s, got: %s", v.Result.D.String(), res.D.String())
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
