// Package ma implements general MA validation functions
// and specifies MA interface that must be implemented
// by every MA type.
package ma

import (
	"eonbot/pkg/exchange"
	"errors"

	"github.com/shopspring/decimal"
)

// MA types
const (
	SMAName = "sma"
	EMAName = "ema"
	WMAName = "wma"
)

var (
	ErrMATypeInvalid = errors.New("MA type is invalid")
	ErrSamePeriodMAs = errors.New("two MAs of the same type cannot be of the same period")
)

// MA is used to perform calculations of specific moving-average.
type MA interface {
	// CandlesCount returns min amount of candles needed to perform
	// MA calculations.
	CandlesCount() int

	// Calc is used to calculate specific MA from candles (open, high,
	// low, close) values.
	Calc(cc []exchange.Candle) (decimal.Decimal, error)

	// CalcDecimal is used to calculate specific MA from provided
	// slice values.
	CalcDecimal(dd []decimal.Decimal) (decimal.Decimal, error)
}

// MAConfig contains settings needed
// to calculate specific MA.
type MAConfig struct {
	// Period specifies how many candles/values are
	// needed to calculate MA value.
	Period int `json:"period"`

	// Price specifies which candle price (Open, High, Low, Close)
	// should be used when calculating MA value.
	Price string `json:"price" conform:"trim,lower"`
}

// validate checks if MAConfig values
// are valid and usable.
func (m *MAConfig) Validate() error {
	if err := PeriodValidation(m.Period); err != nil {
		return err
	}

	if err := exchange.CandlePriceValid(m.Price); err != nil {
		return err
	}

	return nil
}

// PeriodValidation checks if provided period
// is valid (ranging from 1 to 200).
func PeriodValidation(period int) error {
	if period < 1 || period > 200 {
		return errors.New("period must be between 1 and 200 (inclusively)")
	}
	return nil
}

// MATypeValidation checks if provided
// MA type string is valid and existing.
func MATypeValidation(t string) error {
	switch t {
	case SMAName, EMAName, WMAName:
		return nil
	default:
		return ErrMATypeInvalid
	}
}

// TwoMAValidation checks if provided MAs
// require different amount of candles/values.
func TwoMAValidation(t1 string, ma1 MA, t2 string, ma2 MA) error {
	if t1 == t2 && ma1.CandlesCount() == ma2.CandlesCount() {
		return ErrSamePeriodMAs
	}
	return nil
}
