package tools

import (
	"errors"
)

const (
	CalcPercent = "percent"
	CalcUnits   = "units"
	CalcFixed   = "fixed"
)

var (
	ErrCalcInvalid = errors.New("calc is invalid")
)

type Calc struct {
	Type       string `json:"calcType" conform:"trim,lower"`
	allowFixed bool
}

func (c *Calc) AllowFixed() {
	c.allowFixed = true
}

func (c *Calc) Validate() error {
	if err := CalcValidation(c.Type); err != nil {
		return err
	}

	if !c.allowFixed && c.Type == CalcFixed {
		return ErrCalcInvalid
	}
	return nil
}

func CalcValidation(c string) error {
	switch c {
	case CalcPercent, CalcUnits, CalcFixed:
		return nil
	default:
		return ErrCalcInvalid
	}
}
