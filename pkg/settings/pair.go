package settings

import (
	"encoding/json"
	"eonbot/pkg/strategy"
	"errors"
	"fmt"

	"github.com/leebenson/conform"
)

type Pair struct {
	// CandleInterval specifies candles interval type in seconds.
	CandleInterval int `json:"candleInterval"`

	// OrderHistoryDayCount specifies how many days of order history to retrieve
	// from exchange when trying to find buy price.
	OrderHistoryDayCount int `json:"orderHistoryDayCount"`

	// strategies specifies custom strategies names that should be executed on every
	// cycle.
	Strategies []string `json:"strategies" conform:"trim"`

	// CancelOpenOrders specifies whether open orders should be cancelled or not.
	CancelOpenOrders bool `json:"cancelOpenOrders"`

	// OpenOrderLifespan specifies how long should the bot wait (in seconds) til
	// it should cancel an open order. CancelOpenOrders must be set to true.
	OpenOrderLifespan int64 `json:"openOrdersLifespan"`
}

func (p Pair) validate() error {
	if p.CandleInterval <= 0 {
		return errors.New("interval cannot be 0 or less")
	}

	if p.OrderHistoryDayCount < 1 {
		return errors.New("order history start cannot be less than 1")
	}

	if p.Strategies == nil || len(p.Strategies) <= 0 {
		return errors.New("strategies list cannot be empty")
	}

	if err := p.checkDupStrats(); err != nil {
		return err
	}

	if p.CancelOpenOrders {
		if p.OpenOrderLifespan < 10 {
			return errors.New("open orders lifespan cannot be less than 10")
		}
	}

	return nil
}

// checkDupStrats checks if two or more strategies have the same names.
func (p Pair) checkDupStrats() error {
	checked := make([]string, 0)
	for _, strat := range p.Strategies {
		for _, checkedStrat := range checked {
			if checkedStrat == strat {
				return fmt.Errorf("%s strategy is being used more than once in the strategies list", strat)
			}
		}
	}

	return nil
}

// ValidateStrategies checks if strategies' names specified in config are valid
// and point to an existing strategy.
func (p Pair) ValidateStrategies(active map[string]strategy.Strategy) error {
Outer:
	for _, str := range p.Strategies {
		for _, acStr := range active {
			if str == acStr.Name() {
				continue Outer
			}
		}

		return fmt.Errorf("'%s' strategy does not exist", str)
	}

	return nil
}

func (p *Pair) UnmarshalJSON(d []byte) error {
	type TmpPair Pair

	var tmp TmpPair

	if err := json.Unmarshal(d, &tmp); err != nil {
		return err
	}

	*p = Pair(tmp)

	conform.Strings(p)
	return p.validate()
}
