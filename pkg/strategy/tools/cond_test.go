package tools

import (
	"encoding/json"
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCondMatch(t *testing.T) {
	tests := []struct {
		Name   string
		C      Cond
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Result bool
	}{
		{
			Name: "Unsuccessful func call",
			C: Cond{
				C: "",
			},
			Val1:   decimal.RequireFromString("10"),
			Val2:   decimal.RequireFromString("11"),
			Result: false,
		},
		{
			Name: "Successful func call, but result is false",
			C: Cond{
				C: CondEqual,
			},
			Val1:   decimal.RequireFromString("10"),
			Val2:   decimal.RequireFromString("11"),
			Result: false,
		},
		{
			Name: "Successful func call and result is true when val1 is greater than val2",
			C: Cond{
				C: CondAbove,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("10"),
			Result: true,
		},
		{
			Name: "Successful func call and result is true when val1 is greater than or equal than val2",
			C: Cond{
				C: CondAboveOrEqual,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("10"),
			Result: true,
		},
		{
			Name: "Successful func call and result is true when val1 is less than val2",
			C: Cond{
				C: CondBelow,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("12"),
			Result: true,
		},
		{
			Name: "Successful func call and result is true when val1 is less than or equal to val2",
			C: Cond{
				C: CondBelowOrEqual,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("12"),
			Result: true,
		},
		{
			Name: "Successful func call and result is true when val1 is equal to val2",
			C: Cond{
				C: CondEqual,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("11"),
			Result: true,
		},
		{
			Name: "Successful func call and result is true when val1 is above ot below val2",
			C: Cond{
				C: CondAboveOrBelow,
			},
			Val1:   decimal.RequireFromString("11"),
			Val2:   decimal.RequireFromString("10"),
			Result: true,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.C.Match(v.Val1, v.Val2)
			if res != v.Result {
				t.Errorf("incorrect result, expected: %v, got: %v", v.Result, res)
			}
		})
	}
}

func TestCondValidate(t *testing.T) {
	tests := []struct {
		Name        string
		C           Cond
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when no possible conditions found",
			C:           Cond{},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			C: Cond{
				C: CondEqual,
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.C.Validate()
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

func TestCondObjectInit(t *testing.T) {
	tests := []struct {
		Name        string
		C           CondObject
		ShouldError bool
		Offset      int
	}{
		{
			Name:        "Unsuccessful init when offset is invalid",
			ShouldError: true,
			Offset:      -1,
			C:           CondObject{},
		},
		{
			Name:        "Unsuccessful init when object is invalid",
			C:           CondObject{Obj: "test"},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful init when object is set to LastPrice, but ticker prices are not allowed",
			C:           CondObject{Obj: exchange.LastPrice},
			ShouldError: true,
		},
		{
			Name: "Successful init when object is set to LastPrice",
			C: func() CondObject {
				val := CondObject{Obj: exchange.LastPrice}
				val.AllowTickerPrice()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful init when object is set to TickerBaseVolume, but ticker misc properties are not allowed",
			C:           CondObject{Obj: exchange.TickerBaseVolume},
			ShouldError: true,
		},
		{
			Name: "Successful init when object is set to TickerBaseVolume",
			C: func() CondObject {
				val := CondObject{Obj: exchange.TickerBaseVolume}
				val.AllowTickerMiscProp()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful init when object is set to ClosePrice, but candle prices are not allowed",
			C:           CondObject{Obj: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name: "Successful init when object is set to ClosePrice",
			C: func() CondObject {
				val := CondObject{Obj: exchange.ClosePrice}
				val.AllowCandlePrice()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful init when object is set to SMA, but MAs are not allowed",
			C:           CondObject{Obj: ma.SMAName},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful init when object is set to SMA, but the json config is invalid",
			C: func() CondObject {
				val := CondObject{
					Obj:     ma.SMAName,
					ObjConf: json.RawMessage(`{"price":"test"`),
				}
				val.AllowMA()
				return val
			}(),
			ShouldError: true,
		},
		{
			Name: "Unsuccessful init when object is set to SMA, but the price is invalid",
			C: func() CondObject {
				val := CondObject{
					Obj:     ma.SMAName,
					ObjConf: json.RawMessage(`{"price":"test"}`),
				}
				val.AllowMA()
				return val
			}(),
			ShouldError: true,
		},
		{
			Name: "Successful init when object is set to SMA",
			C: func() CondObject {
				val := CondObject{
					Obj:     ma.SMAName,
					ObjConf: json.RawMessage(`{"price":"close", "period":20}`),
				}
				val.AllowMA()
				return val
			}(),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.C.Init(v.Offset)
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

func TestCondObjectValidate(t *testing.T) {
	tests := []struct {
		Name        string
		C           CondObject
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when object is invalid",
			C:           CondObject{},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when object is set to LastPrice, but ticker prices are not allowed",
			C:           CondObject{Obj: exchange.LastPrice},
			ShouldError: true,
		},
		{
			Name: "Successful validation when object is set to LastPrice",
			C: func() CondObject {
				val := CondObject{Obj: exchange.LastPrice}
				val.AllowTickerPrice()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful validation when object is set to TickerBaseVolume, but ticker misc properties are not allowed",
			C:           CondObject{Obj: exchange.TickerBaseVolume},
			ShouldError: true,
		},
		{
			Name: "Successful validation when object is set to LastPrice",
			C: func() CondObject {
				val := CondObject{Obj: exchange.TickerBaseVolume}
				val.AllowTickerMiscProp()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful validation when object is set to ClosePrice, but candle prices are not allowed",
			C:           CondObject{Obj: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name: "Successful validation when object is set to ClosePrice",
			C: func() CondObject {
				val := CondObject{Obj: exchange.ClosePrice}
				val.AllowCandlePrice()
				return val
			}(),
			ShouldError: false,
		},
		{
			Name:        "Unsuccessful validation when object is set to SMA, but MAs are not allowed",
			C:           CondObject{Obj: ma.SMAName},
			ShouldError: true,
		},
		{
			Name: "Successful validation when object is set to SMA",
			C: func() CondObject {
				val := CondObject{
					Obj: ma.SMAName,
				}
				val.AllowMA()
				return val
			}(),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.C.Validate()
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

func TestCondObjectValue(t *testing.T) {
	tests := []struct {
		Name        string
		C           CondObject
		Result      decimal.Decimal
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful value return when retriever is not initialized",
			C:           CondObject{},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful value return when value retriever returns error",
			C: CondObject{
				getData: func(d exchange.Data) (decimal.Decimal, error) {
					return decimal.Zero, errors.New("test")
				},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Unsuccessful value return when value is invalid",
			C: CondObject{
				getData: func(d exchange.Data) (decimal.Decimal, error) {
					return decimal.RequireFromString("0"), nil
				},
			},
			Result:      decimal.Zero,
			ShouldError: true,
		},
		{
			Name: "Successful value return",
			C: CondObject{
				getData: func(d exchange.Data) (decimal.Decimal, error) {
					return decimal.RequireFromString("10"), nil
				},
			},
			Result:      decimal.RequireFromString("10"),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.C.Value(exchange.Data{})
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
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

func TestCondObjectCandlesCount(t *testing.T) {
	tests := []struct {
		Name   string
		C      CondObject
		Result int
	}{
		{
			Name:   "Unsuccessful count return when retriever is not initialized",
			C:      CondObject{},
			Result: 0,
		},
		{
			Name: "Successful count return",
			C: CondObject{
				getCandlesCount: func() int {
					return 10
				},
			},
			Result: 10,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.C.CandlesCount()
			if res != v.Result {
				t.Errorf("incorrect result, expected: %d, got: %d", v.Result, res)
			}
		})
	}
}
