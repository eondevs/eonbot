package tools

import (
	ebMath "eonbot/pkg/math"
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrShiftValInvalid = errors.New("shift val cannot be zero")
)

type Shift struct {
	ShiftVal decimal.Decimal `json:"shiftVal"`
	Calc
}

func (s *Shift) CalcVal(val decimal.Decimal) (res decimal.Decimal) {
	switch s.Calc.Type {
	case CalcUnits:
		return ebMath.UnitsIncrease(val, s.ShiftVal)
	case CalcPercent:
		return ebMath.PercentIncrease(val, s.ShiftVal)
	case CalcFixed:
		if s.Calc.allowFixed {
			return s.ShiftVal
		}
		return decimal.Zero
	default:
		return decimal.Zero
	}
}

func (s *Shift) Validate() error {
	if s.ShiftVal.Equal(decimal.Zero) {
		return ErrShiftValInvalid
	}

	if err := s.Calc.Validate(); err != nil {
		return err
	}

	return nil
}
