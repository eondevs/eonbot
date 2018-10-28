package stoch

import "eonbot/pkg/strategy/indicators/ma"

// StochConfig contains settings needed
// to calculate Stoch.
type StochConfig struct {
	// KPeriod specifies how many candles/values are
	// needed to calculate K line point.
	KPeriod int `json:"KPeriod"`

	// DPeriod specifies how many K values are
	// needed to calculate D line point.
	DPeriod int `json:"DPeriod"`
}

// validate checks if StochConfig values
// are valid and usable.
func (s *StochConfig) Validate() error {
	if err := ma.PeriodValidation(s.KPeriod); err != nil {
		return err
	}

	if err := ma.PeriodValidation(s.DPeriod); err != nil {
		return err
	}

	return nil
}
