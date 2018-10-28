package macd

import (
	"eonbot/pkg/exchange"
	"testing"
)

func TestMACDConfigValidation(t *testing.T) {
	tests := []struct {
		Name        string
		Config      MACDConfig
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when ema1 period is invalid",
			Config:      MACDConfig{EMA1Period: 300, EMA2Period: 10, SignalPeriod: 10, Price: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when ema2 period is invalid",
			Config:      MACDConfig{EMA1Period: 10, EMA2Period: 1000, SignalPeriod: 10, Price: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when signal period is invalid",
			Config:      MACDConfig{EMA1Period: 10, EMA2Period: 10, SignalPeriod: 1000, Price: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when price is invalid",
			Config:      MACDConfig{EMA1Period: 10, EMA2Period: 10, SignalPeriod: 100, Price: "test"},
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			Config:      MACDConfig{EMA1Period: 10, EMA2Period: 10, SignalPeriod: 100, Price: exchange.ClosePrice},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.Config.Validate()
			if v.ShouldError {
				if err == nil {
					t.Error("error expected, but not returned")
				}
			} else {
				if err != nil {
					t.Error("error not expected, but returned:", err)
				}
			}
		})
	}
}
