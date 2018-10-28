package math

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestUnitsIncrease(t *testing.T) {
	tests := []struct {
		Name   string
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Result decimal.Decimal
	}{
		{
			Name:   "Successful calc",
			Val1:   decimal.New(10, 0),
			Val2:   decimal.New(20, 0),
			Result: decimal.New(30, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := UnitsIncrease(v.Val1, v.Val2)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}

func TestPercentIncrease(t *testing.T) {
	tests := []struct {
		Name   string
		Val1   decimal.Decimal
		Val2   decimal.Decimal
		Result decimal.Decimal
	}{
		{
			Name:   "Successful calc",
			Val1:   decimal.New(10, 0),
			Val2:   decimal.New(20, 0),
			Result: decimal.New(12, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := PercentIncrease(v.Val1, v.Val2)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}
