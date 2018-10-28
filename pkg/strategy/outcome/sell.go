package outcome

import (
	"eonbot/pkg/exchange"
	"errors"
)

type Sell struct {
	Price string `json:"price" conform:"trim,lower"`
}

func (s Sell) Validate() error {
	switch s.Price {
	case exchange.LastPrice, exchange.AskPrice, exchange.BidPrice:
		break
	default:
		return errors.New("price type is invalid")
	}
	return nil
}

func (s *Sell) Reset() {}
