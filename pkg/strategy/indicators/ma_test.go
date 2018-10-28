package indicators

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators/ma"
	"testing"
)

func TestNewMA(t *testing.T) {
	tests := []struct {
		Name        string
		MAType      string
		Conf        bool
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful MA creation when MA type is invalid",
			MAType:      "test",
			ShouldError: true,
		},
		{
			Name:   "Successful SMA creation",
			MAType: ma.SMAName,
			Conf:   true,
		},
		{
			Name:   "Successful EMA creation",
			MAType: ma.EMAName,
		},
		{
			Name:   "Successful WMA creation",
			MAType: ma.WMAName,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			var err error
			if v.Conf {
				_, err = NewMAFromConfig(v.MAType, ma.MAConfig{Period: 3, Price: exchange.ClosePrice}, 0)
			} else {
				_, err = NewMA(v.MAType, exchange.ClosePrice, 3, 0)
			}
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
