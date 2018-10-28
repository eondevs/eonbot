package macd

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
)

// MACDConfig contains settings needed
// to calculate MACD.
type MACDConfig struct {
	// EMA1Period specifies how many candles/values are
	// needed to calculate EMA1 point.
	EMA1Period int `json:"ema1Period"`

	// EMA2Period specifies how many candles/values are
	// needed to calculate EMA2 point.
	EMA2Period int `json:"ema2Period"`

	// SignalPeriod specifies how many MACD line values are
	// needed to calculate SignalLine point.
	SignalPeriod int `json:"signalPeriod"`

	// Price specifies which candle price (Open, High, Low, Close)
	// should be used when calculating MACD.
	Price string `json:"price" conform:"trim,lower"`
}

// validate checks if MACDConfig values
// are valid and usable.
func (m *MACDConfig) Validate() error {
	if err := ma.PeriodValidation(m.EMA1Period); err != nil {
		return err
	}

	if err := ma.PeriodValidation(m.EMA2Period); err != nil {
		return err
	}

	if err := ma.PeriodValidation(m.SignalPeriod); err != nil {
		return err
	}

	if err := exchange.CandlePriceValid(m.Price); err != nil {
		return err
	}

	return nil
}
