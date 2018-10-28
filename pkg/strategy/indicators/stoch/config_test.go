package stoch

import "testing"

func TestStochConfigValidation(t *testing.T) {
	tests := []struct {
		Name        string
		Config      StochConfig
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful validation when K period is invalid",
			Config:      StochConfig{KPeriod: 300, DPeriod: 10},
			ShouldError: true,
		},
		{
			Name:        "Unsuccessful validation when D period is invalid",
			Config:      StochConfig{KPeriod: 10, DPeriod: 1000},
			ShouldError: true,
		},
		{
			Name:        "Successful validation",
			Config:      StochConfig{KPeriod: 14, DPeriod: 3},
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
