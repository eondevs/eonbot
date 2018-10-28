package bb

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"

	"github.com/shopspring/decimal"
)

// BBConfig contains settings needed
// to calculate BB.
type BBConfig struct {
	// Period specifies how many candles/values are
	// needed to calculate BB.
	Period int `json:"period"`

	// STDEV specifes value that will be used
	// to multiply standard deviation when calculating
	// Upper/Lower bands.
	STDEV decimal.Decimal `json:"stdev"`

	// Price specifies which candle price (Open, High, Low, Close)
	// should be used when calculating BB.
	Price string `json:"price" conform:"trim,lower"`

	// MAType specifies which MA (SMA, EMA, WMA) should be used
	// as a middle band when calculating BB.
	MAType string `json:"maType" conform:"trim,lower"`
}

// validate checks if BBConfig values
// are valid and usable.
func (b *BBConfig) Validate() error {
	if err := ma.PeriodValidation(b.Period); err != nil {
		return err
	}

	if b.STDEV.LessThanOrEqual(decimal.Zero) {
		return errors.New("STDEV must be a positive value")
	}

	if err := exchange.CandlePriceValid(b.Price); err != nil {
		return err
	}

	if err := ma.MATypeValidation(b.MAType); err != nil {
		return err
	}

	return nil
}
