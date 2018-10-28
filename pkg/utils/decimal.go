package utils

import "github.com/shopspring/decimal"

func RoundByStep(val, step decimal.Decimal, lower bool) decimal.Decimal {
	step = step.Abs()
	if lower {
		return val.Div(step).Floor().Mul(step)
	}
	return val.Div(step).Round(0).Mul(step)
}

func PreventZero(dec decimal.Decimal) decimal.Decimal {
	if dec.IsZero() {
		return decimal.New(1, 0)
	}
	return dec
}
