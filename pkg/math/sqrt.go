package math

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

func Sqrt(base decimal.Decimal) decimal.Decimal {
	val, _ := base.Float64()
	return decimal.RequireFromString(fmt.Sprint(math.Sqrt(val)))
}
