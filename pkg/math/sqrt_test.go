package math

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestSqrt(t *testing.T) {
	tests := []struct {
		Name   string
		Val    decimal.Decimal
		Result decimal.Decimal
	}{
		{
			Name:   "Successful calc",
			Val:    decimal.New(100, 0),
			Result: decimal.New(10, 0),
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res := Sqrt(v.Val)
			if !res.Equal(v.Result) {
				t.Errorf("incorrect result, expected: %s, got: %s", v.Result.String(), res.String())
			}
		})
	}
}
