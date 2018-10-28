package settings

import (
	"encoding/json"
	"eonbot/pkg/asset"
	"errors"
	"fmt"
)

// Bot contains general bot settings (not related to
// exchanges or strategies).
type Bot struct {
	// CycleDelay specifies the amount of time needed
	// to wait between cycles. In seconds.
	CycleDelay int64 `json:"cycleDelay"`

	// ActivePairs specifies which pairs
	// should the bot use in calculations.
	ActivePairs []asset.Pair `json:"activePairs"`

	// StreamCount specifies how many concurrent
	// pairs' streams should be active at once.
	StreamCount int `json:"streamCount"`

	// SideTaskRestarts specifies the amount of times sellAll/cancelAll
	// should be restarted if error occurs.
	SideTaskRestarts int `json:"sideTaskRestarts"`
}

func (b Bot) validate() error {
	if b.CycleDelay < 5 {
		return errors.New("cycle delay cannot be less than 5")
	}

	if b.ActivePairs == nil || len(b.ActivePairs) <= 0 {
		return errors.New("active pairs list cannot be empty")
	}

	if err := b.checkDupPairs(); err != nil {
		return err
	}

	if b.StreamCount < 1 {
		return errors.New("stream count cannot be less than 1")
	}

	if b.SideTaskRestarts < 1 {
		return errors.New("side tasks restarts count cannot be less than 1")
	}

	return nil
}

// checkDupPairs checks if two or more active pairs have the same code.
func (b Bot) checkDupPairs() error {
	checked := make([]asset.Pair, 0)
	for _, pair := range b.ActivePairs {
		for _, checkedPair := range checked {
			if checkedPair.Equal(pair) {
				return fmt.Errorf("%s pair is being used more than once in the active pairs list", pair.String())
			}
		}
	}

	return nil
}

func (b *Bot) UnmarshalJSON(d []byte) error {
	type TmpBot Bot

	var tmp TmpBot

	if err := json.Unmarshal(d, &tmp); err != nil {
		return err
	}

	*b = Bot(tmp)

	return b.validate()
}
