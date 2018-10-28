package simple

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSimpleChangeNew(t *testing.T) {
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

func TestSimpleChangeValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when CondObject has invalid object",
			Settings: settings{
				CondObject: tools.CondObject{
					Obj: "test",
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			Name: "Unsuccessful validation when Cond has no possible conditions",
			Settings: settings{
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Cond: tools.Cond{},
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
			Name: "Unsuccessful validation when Shift has shift val as zero",
			Settings: settings{
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			Name: "Successful validation",
			Settings: settings{
				CondObject: func() tools.CondObject {
					val := tools.CondObject{
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					return val
				}(),
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
				Shift: tools.Shift{
					Calc: tools.Calc{
						Type: tools.CalcUnits,
					},
					ShiftVal: decimal.RequireFromString("10"),
				},
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := SimpleChange{conf: v.Settings}
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

func TestSimpleChangeConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        SimpleChange
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
			Name: "Successful func call when cached val is set to default and result is false",
			Tool: SimpleChange{
				conf: settings{
					Shift: tools.Shift{
						ShiftVal: decimal.RequireFromString("10"),
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
					},
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
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					ChangeValue: decimal.RequireFromString("20.1"),
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
			Name: "Successful func call when cached val is set to non-default and result is false",
			Tool: SimpleChange{
				val: decimal.RequireFromString("101"),
				conf: settings{
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Cond: tools.Cond{C: tools.CondEqual},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					ChangeValue: decimal.New(101, 0),
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
			Name: "Successful func call with fixed shift val",
			Tool: SimpleChange{
				conf: settings{
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Cond: tools.Cond{C: tools.CondEqual},
					Shift: func() tools.Shift {
						val := tools.Shift{
							ShiftVal: decimal.New(100, 0),
							Calc:     tools.Calc{Type: tools.CalcFixed},
						}
						val.AllowFixed()
						return val
					}(),
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					ChangeValue: decimal.New(100, 0),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.New(100, 0),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.New(100, 0)},
			},
			Result:      true,
			ShouldError: false,
		},
		{
			Name: "Successful func call",
			Tool: SimpleChange{
				val: decimal.RequireFromString("101"),
				conf: settings{
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Cond: tools.Cond{C: tools.CondBelow},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					ChangeValue: decimal.New(101, 0),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("10.1"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("10.1")},
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

func TestSimpleChangeCandlesCount(t *testing.T) {
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
						Obj: exchange.LastPrice,
					}
					val.AllowTickerPrice()
					val.Init(0)
					return val
				}(),
			},
			Result: 0,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := SimpleChange{conf: v.Settings}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}

func TestSimpleChangeReset(t *testing.T) {
	s := SimpleChange{val: decimal.New(100, 0)}
	s.Reset()
	assert.Equal(t, decimal.Zero, s.val)
}
