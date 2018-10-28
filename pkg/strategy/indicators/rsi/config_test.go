package rsi

import (
	"eonbot/pkg/exchange"
	"testing"
)

func TestRSIConfigValidation(t *testing.T) {
	tests := []struct {
		Name        string
		Config      RSIConfig
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when period is invalid",
			Config:      RSIConfig{Period: 300, Price: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when price is invalid",
			Config:      RSIConfig{Period: 10, Price: "test"},
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			Config:      RSIConfig{Period: 10, Price: exchange.LowPrice},
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
