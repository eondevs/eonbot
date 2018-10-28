package rsi

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/rsi_mock"
	indiRSI "eonbot/pkg/strategy/indicators/rsi"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestRSINew(t *testing.T) {
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
			Name: "Unsuccessful creation when RSIConfig's price type is invalid",
			Conf: func(v interface{}) error {
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.RSIConfig.Price = exchange.ClosePrice
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

func TestRSIValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when RSIConfig has invalid price type",
			Settings: settings{
				RSIConfig: indiRSI.RSIConfig{
					Price:  "test",
					Period: 10,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Level: func() tools.Level {
					val := tools.Level{
						LevelVal: decimal.RequireFromString("10"),
					}
					val.ZeroToHundred()
					return val
				}(),
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Cond has no possible conditions",
			Settings: settings{
				RSIConfig: indiRSI.RSIConfig{
					Price:  exchange.ClosePrice,
					Period: 10,
				},
				Cond: tools.Cond{},
				Level: func() tools.Level {
					val := tools.Level{
						LevelVal: decimal.RequireFromString("10"),
					}
					val.ZeroToHundred()
					return val
				}(),
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Level has val as zero",
			Settings: settings{
				RSIConfig: indiRSI.RSIConfig{
					Price:  exchange.ClosePrice,
					Period: 10,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Level: func() tools.Level {
					val := tools.Level{
						LevelVal: decimal.Zero,
					}
					val.ZeroToHundred()
					return val
				}(),
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Settings: settings{
				RSIConfig: indiRSI.RSIConfig{
					Price:  exchange.ClosePrice,
					Period: 10,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Level: func() tools.Level {
					val := tools.Level{
						LevelVal: decimal.RequireFromString("10"),
					}
					val.ZeroToHundred()
					return val
				}(),
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := RSI{conf: v.Settings}
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

func TestRSIConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        RSI
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name: "Unsuccessful func call when RSI calc returns an error",
			Tool: RSI{
				rsi: rsi_mock.NewRSIMock(decimal.Zero, errors.New("test"), 1, 1),
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call",
			Tool: RSI{
				rsi: rsi_mock.NewRSIMock(decimal.RequireFromString("10"), nil, 3, 0),
				conf: settings{
					Cond: tools.Cond{
						C: tools.CondEqual,
					},
					Level: tools.Level{
						LevelVal: decimal.RequireFromString("10"),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					RSIVal: decimal.RequireFromString("10"),
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

func TestRSICandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		RSI    indiRSI.RSI
		Result int
	}{
		{
			Name:   "Successful candles count return",
			RSI:    rsi_mock.NewRSIMock(decimal.Zero, nil, 3, 2),
			Result: 8,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := RSI{rsi: v.RSI}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
