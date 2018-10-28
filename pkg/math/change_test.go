package math

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestUnitsChange(t *testing.T) {
	tests := []struct {
		Name   string
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Result decimal.Decimal
	}{
		{
			Name:   "Successful increase calc",
			Val1:   decimal.New(10, 0),
			Val2:   decimal.New(20, 0),
			Result: decimal.New(10, 0),
		},
		{
			Name:   "Successful decrease calc",
			Val1:   decimal.New(20, 0),
			Val2:   decimal.New(10, 0),
			Result: decimal.New(-10, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := UnitsChange(v.Val1, v.Val2)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}

func TestPercentChange(t *testing.T) {
	tests := []struct {
		Name   string
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Result decimal.Decimal
	}{
		{
			Name:   "Successful increase calc",
			Val1:   decimal.New(10, 0),
			Val2:   decimal.New(20, 0),
			Result: decimal.New(100, 0),
		},
		{
			Name:   "Successful decrease calc",
			Val1:   decimal.New(20, 0),
			Val2:   decimal.New(10, 0),
			Result: decimal.New(-50, 0),
		},
		{
			Name:   "Successful increase calc with invalid val1",
			Val1:   decimal.New(0, 0),
			Val2:   decimal.New(10, 0),
			Result: decimal.New(0, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := PercentChange(v.Val1, v.Val2)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}
