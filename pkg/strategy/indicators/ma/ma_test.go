package ma

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/all_mocks/sma_mock"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMAConfigValidation(t *testing.T) {
	tests := []struct {
		Name        string
		Config      MAConfig
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when period is invalid",
			Config:      MAConfig{Period: 300, Price: exchange.ClosePrice},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when price type is invalid",
			Config:      MAConfig{Period: 10, Price: "test"},
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			Config:      MAConfig{Period: 14, Price: exchange.ClosePrice},
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

func TestMATypeValidation(t *testing.T) {
	tests := []struct {
		Name        string
		Type        string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when MA type is invalid",
			Type:        "test",
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			Type:        SMAName,
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := MATypeValidation(v.Type)
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

func TestTwoMAValidation(t *testing.T) {
	tests := []struct {
		Name        string
		MA1         MA
		MA2         MA
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when both MAs are of the same length",
			MA1:         sma_mock.NewSMAMock(decimal.Zero, nil, 10, 12),
			MA2:         sma_mock.NewSMAMock(decimal.Zero, nil, 10, 12),
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			MA1:         sma_mock.NewSMAMock(decimal.Zero, nil, 8, 12),
			MA2:         sma_mock.NewSMAMock(decimal.Zero, nil, 10, 12),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := TwoMAValidation(SMAName, v.MA1, SMAName, v.MA2)
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
