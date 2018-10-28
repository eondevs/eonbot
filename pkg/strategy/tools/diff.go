package tools

import (
	ebMath "eonbot/pkg/math"

	"github.com/shopspring/decimal"
)

type Diff struct {
	Calc
}

func (d *Diff) Validate() error {
	if err := d.Calc.Validate(); err != nil {
		return err
	}

	return nil
}

func (d *Diff) UnitsDiff(val1, val2 decimal.Decimal) decimal.Decimal {
	return ebMath.UnitsChange(val1, val2)
}

func (d *Diff) PercentDiff(val1, val2 decimal.Decimal) decimal.Decimal {
	return ebMath.PercentChange(val1, val2)
}

func (d *Diff) Diff(val1, val2 decimal.Decimal) decimal.Decimal {
	switch d.Calc.Type {
	case CalcUnits:
		return d.UnitsDiff(val1, val2)
	case CalcPercent:
		return d.PercentDiff(val1, val2)
	default:
		return decimal.Zero
	}
}
