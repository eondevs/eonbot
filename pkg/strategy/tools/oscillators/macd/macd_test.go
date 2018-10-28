package macd

import (
	"errors"
	"testing"

	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/macd_mock"
	indiMACD "eonbot/pkg/strategy/indicators/macd"
	"eonbot/pkg/strategy/tools"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMACDNew(t *testing.T) {
	tests := []struct {
		Name        string
		Conf        func(v interface{}) error
		ShouldError bool
	}{
		{
			Name: "Unsuccessful creation when passed function returns error",
			Conf: func(v interface{}) error {
				return errors.New("test")
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful creation when MACDConfig's price type is invalid",
			Conf: func(v interface{}) error {
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.MACDConfig.Price = exchange.ClosePrice
				return nil
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			_, err := New(v.Conf)
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

func TestMACDValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when MACDConfig has invalid price type",
			Settings: settings{
				MACDConfig: indiMACD.MACDConfig{
					EMA1Period:   20,
					EMA2Period:   21,
					SignalPeriod: 2,
					Price:        "test",
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Cond has no possible conditions",
			Settings: settings{
				MACDConfig: indiMACD.MACDConfig{
					EMA1Period:   20,
					EMA2Period:   21,
					SignalPeriod: 2,
					Price:        exchange.ClosePrice,
				},
				Cond: tools.Cond{},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Diff has invalid calc type",
			Settings: settings{
				MACDConfig: indiMACD.MACDConfig{
					EMA1Period:   20,
					EMA2Period:   21,
					SignalPeriod: 2,
					Price:        exchange.ClosePrice,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: "test",
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Settings: settings{
				MACDConfig: indiMACD.MACDConfig{
					EMA1Period:   20,
					EMA2Period:   21,
					SignalPeriod: 2,
					Price:        exchange.ClosePrice,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := MACD{conf: v.Settings}
			err := obj.Validate()
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

func TestMACDConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        MACD
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name: "Unsuccessful func call when MACD calc returns an error",
			Tool: MACD{
				macd: macd_mock.NewMACDMock(indiMACD.MACDInfo{}, errors.New("test"), 1, 1, 1, 0),
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call",
			Tool: MACD{
				macd: macd_mock.NewMACDMock(indiMACD.MACDInfo{
					MACDLine:   decimal.RequireFromString("9"),
					SignalLine: decimal.RequireFromString("10"),
				}, nil, 1, 1, 1, 0),
				conf: settings{
					Differ: decimal.RequireFromString("-1"),
					Cond: tools.Cond{
						C: tools.CondEqual,
					},
					Diff: tools.Diff{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					Diff: decimal.RequireFromString("-1"),
					MACDInfo: indiMACD.MACDInfo{
						MACDLine:   decimal.RequireFromString("9"),
						SignalLine: decimal.RequireFromString("10"),
					},
				},
			},
			Result:      true,
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.Tool.ConditionsMet(v.Data)
			assert.Equal(t, v.Result, res)
			snap := v.Tool.Snapshot()
			assert.Equal(t, v.Snapshot, snap)
			if v.ShouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestMACDCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		MACD   indiMACD.MACD
		Result int
	}{
		{
			Name:   "Successful candles count return",
			MACD:   macd_mock.NewMACDMock(indiMACD.MACDInfo{}, nil, 2, 3, 2, 1),
			Result: 11,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := MACD{macd: v.MACD}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
