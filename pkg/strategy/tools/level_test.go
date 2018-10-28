package tools

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestLevelValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Level       Level
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when level val is out of bounds",
			Level: func() Level {
				val := Level{LevelVal: decimal.New(-1, 0)}
				val.ZeroToHundred()
				return val
			}(),
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			Level: func() Level {
				val := Level{LevelVal: decimal.New(10, 0)}
				val.ZeroToHundred()
				return val
			}(),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.Level.Validate()
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
