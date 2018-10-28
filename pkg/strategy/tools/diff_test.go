package tools

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDiffValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Diff        Diff
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when Calc type is invalid",
			Diff: Diff{
				Calc: Calc{
					Type: "test",
				},
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Diff: Diff{
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
			err := v.Diff.Validate()
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

func TestDiffDiff(t *testing.T) {
	tests := []struct {
		Name   string
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Diff   Diff
		Result decimal.Decimal
	}{
		{
			Name:   "Unsuccessful diff calculation when Calc type is invalid",
			Diff:   Diff{Calc: Calc{Type: "test"}},
			Val1:   decimal.RequireFromString("10"),
			Val2:   decimal.RequireFromString("10"),
			Result: decimal.Zero,
		},
		{
			Name:   "Successful diff calculation when Calc type is set to percent",
			Diff:   Diff{Calc: Calc{Type: CalcPercent}},
			Val1:   decimal.RequireFromString("10"),
			Val2:   decimal.RequireFromString("20"),
			Result: decimal.RequireFromString("100"),
		},
		{
			Name:   "Successful diff calculation when Calc type is set to units",
			Diff:   Diff{Calc: Calc{Type: CalcUnits}},
			Val1:   decimal.RequireFromString("132"),
			Val2:   decimal.RequireFromString("258"),
			Result: decimal.RequireFromString("126"),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := v.Diff.Diff(v.Val1, v.Val2)
			if !res.Round(1).Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}
