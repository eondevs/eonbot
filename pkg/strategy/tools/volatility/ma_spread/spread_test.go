package ma_spread

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/sma_mock"
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMASpreadNew(t *testing.T) {
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
			Name: "Unsuccessful creation when MAConfig1 price type is invalid",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.MA1.Type = ma.SMAName
				val.MA1.Price = "test"
				val.MA2.Type = ma.SMAName
				val.MA2.Price = exchange.ClosePrice
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful creation when MAConfig2 price type is invalid",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.MA1.Type = ma.SMAName
				val.MA1.Price = exchange.ClosePrice
				val.MA2.Type = ma.SMAName
				val.MA2.Price = "test"
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.MA1.Type = ma.SMAName
				val.MA1.Price = exchange.ClosePrice
				val.MA2.Type = ma.SMAName
				val.MA2.Price = exchange.ClosePrice
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

func TestMASpreadValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when base ma index is invalid",
			Settings: settings{
				BaseMA: 0,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when MA1 Type is invalid",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: "test",
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when MA2 Type is invalid",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: "test",
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when ma1 has invalid price type",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  "test",
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when ma2 has invalid price type",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  "test",
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when both MAs require the same candles count",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Diff has invalid calc type",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: "test",
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Cond has no possible conditions",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{},
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Settings: settings{
				BaseMA: 1,
				MA1: maConf{
					Type: ma.EMAName,
					MAConfig: ma.MAConfig{
						Period: 10,
						Price:  exchange.ClosePrice,
					},
				},
				MA2: maConf{
					Type: ma.SMAName,
					MAConfig: ma.MAConfig{
						Period: 11,
						Price:  exchange.ClosePrice,
					},
				},
				Diff: tools.Diff{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := MASpread{
				ma1:  sma_mock.NewSMAMock(decimal.Zero, nil, v.Settings.MA1.Period, 0),
				ma2:  sma_mock.NewSMAMock(decimal.Zero, nil, v.Settings.MA2.Period, 0),
				conf: v.Settings,
			}
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

func TestMASpreadConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        MASpread
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name: "Unsuccessful func call when ma1 calc returns an error",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.Zero, errors.New("test"), 3, 0),
				ma2: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 3, 0),
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when ma2 calc returns an error",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 3, 0),
				ma2: sma_mock.NewSMAMock(decimal.Zero, errors.New("test"), 3, 0),
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when base index is invalid",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.RequireFromString("15"), nil, 6, 0),
				ma2: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 3, 0),
				conf: settings{
					Spread: decimal.RequireFromString("-5"),
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
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call when ma1 is the base one",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.RequireFromString("15"), nil, 6, 0),
				ma2: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 3, 0),
				conf: settings{
					BaseMA: 1,
					Spread: decimal.RequireFromString("-5"),
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
					Spread: decimal.RequireFromString("-5"),
					MA1:    decimal.RequireFromString("15"),
					MA2:    decimal.RequireFromString("10"),
				},
			},
			Result:      true,
			ShouldError: false,
		},
		{
			Name: "Successful func call when ma2 is the base one",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 3, 0),
				ma2: sma_mock.NewSMAMock(decimal.RequireFromString("15"), nil, 6, 0),
				conf: settings{
					BaseMA: 2,
					Spread: decimal.RequireFromString("-5"),
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
					Spread: decimal.RequireFromString("-5"),
					MA1:    decimal.RequireFromString("10"),
					MA2:    decimal.RequireFromString("15"),
				},
			},
			Result:      true,
			ShouldError: false,
		},
		{
			Name: "Successful func call when base MA1 is above the other",
			Tool: MASpread{
				ma1: sma_mock.NewSMAMock(decimal.RequireFromString("15"), nil, 3, 0),
				ma2: sma_mock.NewSMAMock(decimal.RequireFromString("10"), nil, 6, 0),
				conf: settings{
					BaseMA: 2,
					Spread: decimal.RequireFromString("5"),
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
					Spread: decimal.RequireFromString("5"),
					MA1:    decimal.RequireFromString("15"),
					MA2:    decimal.RequireFromString("10"),
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

func TestMASpreadCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		MA1    ma.MA
		MA2    ma.MA
		Result int
	}{
		{
			Name:   "Successful candles count return when MA1 requires more candles",
			MA1:    sma_mock.NewSMAMock(decimal.Zero, nil, 8, 0),
			MA2:    sma_mock.NewSMAMock(decimal.Zero, nil, 5, 0),
			Result: 8,
		},
		{
			Name:   "Successful candles count return when MA2 requires more candles",
			MA1:    sma_mock.NewSMAMock(decimal.Zero, nil, 5, 0),
			MA2:    sma_mock.NewSMAMock(decimal.Zero, nil, 7, 0),
			Result: 7,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := MASpread{ma1: v.MA1, ma2: v.MA2}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
