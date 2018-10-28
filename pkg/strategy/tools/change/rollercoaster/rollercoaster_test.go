package rollercoaster

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestRollerCoasterNew(t *testing.T) {
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
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.CondObject.Obj = exchange.LastPrice
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

func TestRollerCoasterValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when point type is invalid",
			Settings: settings{
				PointType: "test",
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("-10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when CondObject has invalid object",
			Settings: settings{
				PointType: highestPoint,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: "test",
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("-10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Shift has shift val as zero",
			Settings: settings{
				PointType: highestPoint,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.Zero,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when point type is set to highest and shift val is positive",
			Settings: settings{
				PointType: highestPoint,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when point type is set to lowest and shift val is negative",
			Settings: settings{
				PointType: lowestPoint,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("-10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Settings: settings{
				PointType: highestPoint,
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("-10"),
				},
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := RollerCoaster{conf: v.Settings}
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

func TestRollerCoasterConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        RollerCoaster
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful func call when CondObject is not initialized",
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when point type is invalid",
			Tool: RollerCoaster{
				pointVal: decimal.RequireFromString("10"),
				conf: settings{
					PointType: "test",
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("10.1")},
			},
			ShouldError: true,
		},
		{
			Name: "Successful func call when cached point val is set to default and result is false",
			Tool: RollerCoaster{
				conf: settings{
					PointType: highestPoint,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.New(-1, 0),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					PointVal:        decimal.RequireFromString("10.1"),
					ShiftedPointVal: decimal.RequireFromString("9.1"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("10.1"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("10.1")},
			},
			Result: false,
		},
		{
			Name: "Successful func call when cached highest point val is set to non-default and new val is above",
			Tool: RollerCoaster{
				pointVal: decimal.RequireFromString("10"),
				conf: settings{
					PointType: highestPoint,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.New(-1, 0),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					PointVal:        decimal.RequireFromString("101"),
					ShiftedPointVal: decimal.RequireFromString("100"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("101"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("101")},
			},
			Result: false,
		},
		{
			Name: "Successful func call when cached highest point val is set to non-default and new val is below",
			Tool: RollerCoaster{
				pointVal: decimal.RequireFromString("10"),
				conf: settings{
					PointType: highestPoint,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.New(-1, 0),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					PointVal:        decimal.RequireFromString("10"),
					ShiftedPointVal: decimal.RequireFromString("9"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("9"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("9")},
			},
			Result: true,
		},
		{
			Name: "Successful func call when cached lowest point val is set to non-default and new val is below",
			Tool: RollerCoaster{
				pointVal: decimal.RequireFromString("10"),
				conf: settings{
					PointType: lowestPoint,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.New(1, 0),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					PointVal:        decimal.RequireFromString("9"),
					ShiftedPointVal: decimal.RequireFromString("10"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("9"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("9")},
			},
			Result: false,
		},
		{
			Name: "Successful func call when cached lowest point val is set to non-default and new val is above",
			Tool: RollerCoaster{
				pointVal: decimal.RequireFromString("10"),
				conf: settings{
					PointType: lowestPoint,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.New(1, 0),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					PointVal:        decimal.RequireFromString("10"),
					ShiftedPointVal: decimal.RequireFromString("11"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("11"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("11")},
			},
			Result: true,
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

func TestRollerCoasterCandlesCount(t *testing.T) {
	tests := []struct {
		Name     string
		Settings settings
		Result   int
	}{
		{
			Name: "Successful candles count return",
			Settings: settings{
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.ClosePrice,
					}
					val.AllowCandlePrice()
					val.Init(0)
					return val
				}(),
			},
			Result: 1,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := RollerCoaster{conf: v.Settings}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestRollerCoasterReset(t *testing.T) {
	r := RollerCoaster{pointVal: decimal.New(100, 0)}
	r.Reset()
	assert.Equal(t, decimal.Zero, r.pointVal)
}
