package tools

import (
	"testing"
)

func TestCalcValidate(t *testing.T) {
	tests := []struct {
		Name        string
		C           Calc
		ShouldError bool
	}{
		{
			Name: "Unsuccessful validation when calc type is invalid",
			C: Calc{
				Type: "test",
			},
			ShouldError: true,
		},
		{
			Name: "Unsuccessful validation when calc type is fixed, but it is not allowed",
			C: Calc{
				Type: CalcFixed,
			},
			ShouldError: true,
		},
		{
			Name: "Successful validation",
			C: func() Calc {
				val := Calc{
					Type: CalcFixed,
				}
				val.AllowFixed()
				return val
			}(),
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			err := v.C.Validate()
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
