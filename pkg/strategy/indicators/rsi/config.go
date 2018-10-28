package rsi

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
)

// RSIConfig contains settings needed
// to calculate RSI.
type RSIConfig struct {
	// Period specifies how many candles/values are
	// needed to calculate RSI.
	Period int `json:"period"`

	// Price specifies which candle price (Open, High, Low, Close)
	// should be used when calculating RSI.
	Price string `json:"price" conform:"trim,lower"`
}

// validate checks if RSIConfig values
// are valid and usable.
func (r *RSIConfig) Validate() error {
	if err := ma.PeriodValidation(r.Period); err != nil {
		return err
	}

	if err := exchange.CandlePriceValid(r.Price); err != nil {
		return err
	}

	return nil
}
