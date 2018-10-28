package stoch

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/stoch_mock"
	indiStoch "eonbot/pkg/strategy/indicators/stoch"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestStochNew(t *testing.T) {
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
			Name: "Successful creation",
			Conf: func(v interface{}) error {
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

func TestStochValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when StochConfig has invalid period",
			Settings: settings{
				StochConfig: indiStoch.StochConfig{
					KPeriod: 300,
					DPeriod: 10,
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
				StochConfig: indiStoch.StochConfig{
					KPeriod: 10,
					DPeriod: 10,
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
				StochConfig: indiStoch.StochConfig{
					KPeriod: 10,
					DPeriod: 10,
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
				StochConfig: indiStoch.StochConfig{
					KPeriod: 10,
					DPeriod: 10,
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
			obj := Stoch{conf: v.Settings}
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

func TestStochConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        Stoch
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name: "Unsuccessful func call when Stoch calc returns an error",
			Tool: Stoch{
				stoch: stoch_mock.NewStochMock(indiStoch.StochInfo{}, errors.New("test"), 10, 3, 0),
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call",
			Tool: Stoch{
				stoch: stoch_mock.NewStochMock(indiStoch.StochInfo{D: decimal.RequireFromString("10")}, nil, 10, 3, 0),
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
					StochInfo: indiStoch.StochInfo{D: decimal.RequireFromString("10")},
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

func TestStochCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		Stoch  indiStoch.Stoch
		Result int
	}{
		{
			Name:   "Successful candles count return",
			Stoch:  stoch_mock.NewStochMock(indiStoch.StochInfo{}, nil, 14, 3, 1),
			Result: 17,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := Stoch{stoch: v.Stoch}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
