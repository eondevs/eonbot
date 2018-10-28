package indicators

import (
	"eonbot/pkg/strategy/indicators/ma"
	"eonbot/pkg/strategy/indicators/ma/ema"
	"eonbot/pkg/strategy/indicators/ma/sma"
	"eonbot/pkg/strategy/indicators/ma/wma"
)

// NewMA creates new MA by specified maType.
func NewMA(maType, price string, period, offset int) (ma.MA, error) {
	switch maType {
	case ma.SMAName:
		return sma.New(period, offset, price)
	case ma.EMAName:
		return ema.New(period, offset, price)
	case ma.WMAName:
		return wma.New(period, offset, price)
	default:
		return nil, ma.ErrMATypeInvalid
	}
}

// NewMAFromConfig creates new MA, just like NewMA, but
// from MAConfig.
func NewMAFromConfig(maType string, conf ma.MAConfig, offset int) (ma.MA, error) {
	return NewMA(maType, conf.Price, conf.Period, offset)
}
