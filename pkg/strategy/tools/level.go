package tools

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Level struct {
	LevelVal decimal.Decimal `json:"levelVal"`
	min      decimal.Decimal
	max      decimal.Decimal
}

func (l *Level) Init(min, max decimal.Decimal) {
	l.min = min
	l.max = max
}

func (l *Level) ZeroToHundred() {
	l.Init(decimal.Zero, decimal.New(100, 0))
}

func (l *Level) Validate() error {
	if l.LevelVal.GreaterThan(l.max) || l.LevelVal.LessThanOrEqual(l.min) {
		return fmt.Errorf("level value must be a value ranging from %s to %s", l.min.String(), l.max.String())
	}

	return nil
}
