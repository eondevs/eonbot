package math

import "github.com/shopspring/decimal"

// UnitsIncrease increases val1 by val2.
func UnitsIncrease(val1, val2 decimal.Decimal) decimal.Decimal {
	return val1.Add(val2)
}

// PercentIncrease increases val1 by val2 percent value.
func PercentIncrease(val1, val2 decimal.Decimal) decimal.Decimal {
	return val1.Mul(decimal.New(1, 0).Add(val2.Div(decimal.New(100, 0))))
}
