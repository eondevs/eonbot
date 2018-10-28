package utils

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoundByStep(t *testing.T) {
	assert.Equal(t, "0.005", RoundByStep(decimal.RequireFromString("0.00346643"), decimal.RequireFromString("-0.005"), false).String())
	assert.Equal(t, "0.005", RoundByStep(decimal.RequireFromString("0.00346643"), decimal.RequireFromString("0.005"), false).String())
	assert.Equal(t, "0.005", RoundByStep(decimal.RequireFromString("0.005"), decimal.RequireFromString("0.001"), false).String())
	assert.Equal(t, "0.001", RoundByStep(decimal.RequireFromString("0.0005"), decimal.RequireFromString("0.001"), false).String())
	assert.Equal(t, "0.004", RoundByStep(decimal.RequireFromString("0.00346643"), decimal.RequireFromString("0.002"), false).String())
	assert.Equal(t, "0.003", RoundByStep(decimal.RequireFromString("0.00346643"), decimal.RequireFromString("0.003"), false).String())

	// from examples
	assert.Equal(t, "0.0005", RoundByStep(decimal.RequireFromString("0.000355666"), decimal.RequireFromString("0.0005"), false).String())
	assert.Equal(t, "0", RoundByStep(decimal.RequireFromString("0.000355666"), decimal.RequireFromString("0.0005"), true).String())
	assert.Equal(t, "0.0004", RoundByStep(decimal.RequireFromString("0.000355666"), decimal.RequireFromString("0.0001"), false).String())
	assert.Equal(t, "0.0003", RoundByStep(decimal.RequireFromString("0.000355666"), decimal.RequireFromString("0.0001"), true).String())
}

func TestPreventZero(t *testing.T) {
	assert.Equal(t, "1", PreventZero(decimal.Zero).String())
	assert.Equal(t, "3", PreventZero(decimal.RequireFromString("3")).String())
}
