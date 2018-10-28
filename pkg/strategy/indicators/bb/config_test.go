package bb

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"testing"

	"github.com/shopspring/decimal"
)

func TestBBConfigValidate(t *testing.T) {
	tests := []struct {
		Name string
		BBConfig
		ShouldError bool
	}{
		{
			Name: "Error on invalid period",
			BBConfig: BBConfig{
				Period: 201,
				STDEV:  decimal.New(2, 0),
				Price:  exchange.ClosePrice,
				MAType: ma.SMAName,
			},
			ShouldError: true,
		},
		{
			Name: "Error on invalid STDEV",
			BBConfig: BBConfig{
				Period: 200,
				STDEV:  decimal.New(0, 0),
				Price:  exchange.ClosePrice,
				MAType: ma.SMAName,
			},
			ShouldError: true,
		},
		{
			Name: "Error on invalid price type",
			BBConfig: BBConfig{
				Period: 200,
				STDEV:  decimal.New(2, 0),
				Price:  "test",
				MAType: ma.SMAName,
			},
			ShouldError: true,
		},
		{
			Name: "Error on invalid MA type",
			BBConfig: BBConfig{
				Period: 200,
				STDEV:  decimal.New(2, 0),
				Price:  exchange.ClosePrice,
				MAType: "test",
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			BBConfig: BBConfig{
				Period: 200,
				STDEV:  decimal.New(2, 0),
				Price:  exchange.ClosePrice,
				MAType: ma.SMAName,
			},
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.BBConfig.Validate()
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
