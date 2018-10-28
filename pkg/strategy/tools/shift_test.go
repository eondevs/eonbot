package tools

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestShiftValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Shift       Shift
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when shift val is zero",
			Shift: Shift{
				ShiftVal: decimal.Zero,
				Calc: Calc{
					Type: CalcUnits,
				},
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when calc type is invalid",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: Calc{
					Type: "test",
				},
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: Calc{
					Type: CalcUnits,
				},
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.Shift.Validate()
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

func TestShiftCalcVal(t *testing.T) {
	tests := []struct {
		Name   string
		Val    decimal.Decimal
		Shift  Shift
		Result decimal.Decimal
	}{
		{
			Name: "Unsuccessful func call when Calc type is invalid",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: Calc{
					Type: "test",
				},
			},
			Val:    decimal.RequireFromString("5"),
			Result: decimal.Zero,
		},
		{
			Name: "Unsuccessful func call when Calc type is fixed, but it is not allowed",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: Calc{
					Type: CalcFixed,
				},
			},
			Val:    decimal.RequireFromString("5"),
			Result: decimal.Zero,
		},
		{
			Name: "Successful func call when Calc type is fixed",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: func() Calc {
					val := Calc{
						Type: CalcFixed,
					}
					val.AllowFixed()
					return val
				}(),
			},
			Val:    decimal.RequireFromString("5"),
			Result: decimal.RequireFromString("10"),
		},
		{
			Name: "Successful func call when Calc type is units",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: func() Calc {
					val := Calc{
						Type: CalcUnits,
					}
					val.AllowFixed()
					return val
				}(),
			},
			Val:    decimal.RequireFromString("5"),
			Result: decimal.RequireFromString("15"),
		},
		{
			Name: "Successful func call when Calc type is percent",
			Shift: Shift{
				ShiftVal: decimal.RequireFromString("10"),
				Calc: func() Calc {
					val := Calc{
						Type: CalcPercent,
					}
					return val
				}(),
			},
			Val:    decimal.RequireFromString("5"),
			Result: decimal.RequireFromString("5.5"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.Shift.CalcVal(v.Val)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}
