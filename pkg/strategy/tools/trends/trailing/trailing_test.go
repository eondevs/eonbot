package trailing

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTrailingTrendsNew(t *testing.T) {
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
			Name: "Unsuccessful creation when CondObject has invalid object",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.CondObject.Obj = "test"
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful creation when back index is invalid",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.CondObject.Obj = exchange.HighPrice
				val.BackIndex = -1
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.CondObject.Obj = exchange.HighPrice
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

func TestTrailingTrendsValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when back index is invalid",
			Settings: settings{
				BackIndex: 0,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					return val
				}(),
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
			Name: "Unsuccessful validation when CondObject has invalid object",
			Settings: settings{
				BackIndex: 10,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: "test",
					}
					val.AllowCandlePrice()
					return val
				}(),
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
				BackIndex: 10,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					return val
				}(),
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
			Name: "Unsuccessful validation when Diff calc type is invalid",
			Settings: settings{
				BackIndex: 10,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					return val
				}(),
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
				BackIndex: 10,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					return val
				}(),
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
			obj := TrailingTrends{conf: v.Settings}
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

func TestTrailingTrendsConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        TrailingTrends
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name: "Unsuccessful func call when CondObject1 has invalid object",
			Tool: TrailingTrends{
				leadObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: "test",
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
				backObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
			},
			Data: exchange.Data{
				Candles: []exchange.Candle{
					{Close: decimal.RequireFromString("10")},
				},
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when CondObject2 has invalid object",
			Tool: TrailingTrends{
				leadObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
				backObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: "test",
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
			},
			Data: exchange.Data{
				Candles: []exchange.Candle{
					{Close: decimal.RequireFromString("10")},
				},
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call",
			Tool: TrailingTrends{
				leadObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
				backObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(1)
					return val
				}(),
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
					Diff:    decimal.RequireFromString("-1"),
					LeadObj: decimal.RequireFromString("9"),
					BackObj: decimal.RequireFromString("10"),
				},
			},
			Data: exchange.Data{
				Candles: []exchange.Candle{
					{Close: decimal.RequireFromString("10")},
					{Close: decimal.RequireFromString("9")},
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

func TestTrailingTrendsCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		Tool   TrailingTrends
		Result int
	}{
		{
			Name: "Successful candles count return",
			Tool: TrailingTrends{
				backObj: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(4)
					return val
				}(),
			},
			Result: 5,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.Tool.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
