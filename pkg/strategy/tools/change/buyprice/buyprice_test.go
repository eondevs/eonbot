package buyprice

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBuyPriceNew(t *testing.T) {
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
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBuyPriceValidate(t *testing.T) {
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
			obj := BuyPrice{conf: v.Settings}
			err := obj.Validate()
			if v.ShouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBuyPriceConditionsMet(t *testing.T) {
	tests := []struct {
		Name        string
		Tool        BuyPrice
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
			Name: "Successful func call when BuyPrice is set to default",
			Tool: BuyPrice{
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
			Data: exchange.Data{
				Ticker: exchange.TickerData{
					LastPrice: decimal.RequireFromString("10"),
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("10"),
					},
				},
			},
			Result:      true,
			ShouldError: false,
		},
		{
			Name: "Successful func call when BuyPrice is not set to default and conditions do not match",
			Tool: BuyPrice{
				conf: settings{
					Shift: tools.Shift{
						ShiftVal: decimal.RequireFromString("-1"),
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
					Cond: tools.Cond{
						C: tools.CondEqual,
					},
				},
			},
			Data: exchange.Data{
				BuyPrice: decimal.RequireFromString("10"),
				Ticker: exchange.TickerData{
					LastPrice: decimal.RequireFromString("10"),
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: false,
				Data: snapshot{
					BuyPrice:        decimal.RequireFromString("10"),
					ShiftedBuyPrice: decimal.RequireFromString("9"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("10"),
					},
				},
			},
			Result:      false,
			ShouldError: false,
		},
		{
			Name: "Successful func call when BuyPrice is not set to default and conditions match",
			Tool: BuyPrice{
				conf: settings{
					Shift: tools.Shift{
						ShiftVal: decimal.RequireFromString("-1"),
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
					Cond: tools.Cond{
						C: tools.CondEqual,
					},
				},
			},
			Data: exchange.Data{
				BuyPrice: decimal.RequireFromString("10"),
				Ticker: exchange.TickerData{
					LastPrice: decimal.RequireFromString("9"),
				},
			},
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data: snapshot{
					BuyPrice:        decimal.RequireFromString("10"),
					ShiftedBuyPrice: decimal.RequireFromString("9"),
					CondObjectSnapshot: tools.CondObjectSnapshot{
						Value: decimal.RequireFromString("9"),
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

func TestBuyPriceCandlesCount(t *testing.T) {
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
			obj := BuyPrice{conf: v.Settings}
			res := obj.CandlesCount()
			assert.Equal(t, v.Result, res)
		})
	}
}
