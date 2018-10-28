package math

import (
	"github.com/shopspring/decimal"
)

// UnitsChange calculates change from val1 to val2
// in normal units. val1 should be the 'previous' value
// and val2 should be the 'current' value.
// If value is positive - value increased,
// if negative - reduced (from val1).
// UnitsChange calculation:
// change = val1 - val2
func UnitsChange(val1, val2 decimal.Decimal) decimal.Decimal {
	return val2.Sub(val1)
}

// PercentChange calculates change from val1 to val2
// in percentage. val1 should be the 'previous' value
// and val2 should be the 'current' value.
// If value is positive - value increased,
// if negative - reduced (from val1).
// PercentChange calculation:
// change = (val2 - val1) / val1 x 100
func PercentChange(val1, val2 decimal.Decimal) decimal.Decimal {
	if val1.Equal(decimal.Zero) {
		return decimal.Zero
	}

	return val2.Sub(val1).Div(val1).Mul(decimal.New(100, 0))
}
