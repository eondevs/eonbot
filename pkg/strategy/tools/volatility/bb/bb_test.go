package bb

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/bb_mock"
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	indiBB "eonbot/pkg/strategy/indicators/bb"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBBNew(t *testing.T) {
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
			Name: "Unsuccessful creation when BBConfig's price type is invalid",
			Conf: func(v interface{}) error {
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful creation when CondObject has invalid object",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.BBConfig.MAType = ma.SMAName
				val.BBConfig.Price = exchange.ClosePrice
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "Successful creation",
			Conf: func(v interface{}) error {
				val := v.(*settings)
				val.CondObject.Obj = exchange.LastPrice
				val.BBConfig.MAType = ma.SMAName
				val.BBConfig.Price = exchange.ClosePrice
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

func TestBBValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Settings    settings
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when band type is invalid",
			Settings: settings{
				Band: "test",
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  exchange.ClosePrice,
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			Name: "Unsuccessful validation when BBConfig's price type is invalid",
			Settings: settings{
				Band: bandUpper,
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  "test",
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			Name: "Unsuccessful validation when Cond has no possible conditions",
			Settings: settings{
				Band: bandUpper,
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  exchange.ClosePrice,
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{},
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
			Name: "Unsuccessful validation when CondObject has invalid object",
			Settings: settings{
				Band: bandUpper,
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  exchange.ClosePrice,
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
					ShiftVal: decimal.RequireFromString("10"),
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when Shift has shift val as zero",
			Settings: settings{
				Band: bandUpper,
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  exchange.ClosePrice,
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			Name: "Successful validation",
			Settings: settings{
				Band: bandUpper,
				BBConfig: indiBB.BBConfig{
					Period: 10,
					STDEV:  decimal.New(2, 0),
					Price:  exchange.ClosePrice,
					MAType: ma.SMAName,
				},
				Cond: tools.Cond{
					C: tools.CondAbove,
				},
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
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := BB{conf: v.Settings}
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

func TestBBConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        BB
		Data        exchange.Data
		Snapshot    tools.Snapshot
		Result      bool
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful func call when BB calc returns error",
			Tool:        BB{bb: bb_mock.NewBBMock(indiBB.BBInfo{}, errors.New("test"), 1, 1)},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when band type is invalid",
			Tool: BB{
				bb:   bb_mock.NewBBMock(indiBB.BBInfo{}, nil, 1, 1),
				conf: settings{Band: "test"},
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful func call when CondObject has invalid object",
			Tool: BB{
				bb: bb_mock.NewBBMock(indiBB.BBInfo{}, nil, 1, 1),
				conf: settings{
					Band: bandLower,
					CondObject: tools.CondObject{
						Obj: "test",
					},
				},
			},
			Result:      false,
			ShouldError: true,
		},
		{
			Name: "Successful func call",
			Tool: BB{
				bb: bb_mock.NewBBMock(indiBB.BBInfo{
					Upper:  decimal.RequireFromString("10"),
					Middle: decimal.RequireFromString("7"),
					Lower:  decimal.RequireFromString("4"),
				}, nil, 1, 1),
				conf: settings{
					Band: bandUpper,
					CondObject: func() tools.CondObject {
						val := tools.CondObject{
							Obj: exchange.LastPrice,
						}
						val.AllowTickerPrice()
						val.Init(0)
						return val
					}(),
					Cond: tools.Cond{
						C: tools.CondEqual,
					},
					Shift: tools.Shift{
						Calc: tools.Calc{
							Type: tools.CalcUnits,
						},
						ShiftVal: decimal.RequireFromString("2"),
					},
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					ShiftedBand: decimal.RequireFromString("12"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("12"),
					},
					BBInfo: indiBB.BBInfo{
						Upper:  decimal.RequireFromString("10"),
						Middle: decimal.RequireFromString("7"),
						Lower:  decimal.RequireFromString("4"),
					},
				},
			},
			Data: exchange.Data{
				Ticker: exchange.TickerData{
					LastPrice: decimal.RequireFromString("12"),
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

func TestBBCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		BB     indiBB.BB
		Result int
	}{
		{
			Name:   "Successful candles count return",
			BB:     bb_mock.NewBBMock(indiBB.BBInfo{}, nil, 3, 1),
			Result: 4,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			obj := BB{bb: v.BB}
			res := obj.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
