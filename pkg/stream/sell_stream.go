package stream

import (
	"eonbot/pkg/exchange"
)

// Sell checks if base asset is ready to be sold, if it is,
// it places a sell order with ticker's bid price to fill
// the order as soon as possible.
// If another sell order is open for this pair, it will
// be cancelled.
func (s *Stream) Sell(bal BalancesPair) error {
	// retrieve ticker data
	ticker, err := s.Exchange.GetTicker(s.Pair)
	if err != nil {
		return s.prepError(err)
	}

	// make final checks and apply final changes (number steps, precision rounding).
	rate, amount, err := s.Pair.Transaction(ticker.BidPrice, bal.Base)
	if err != nil {
		// no need to check for error here, if some
		//conditions are not met, don't
		// place an order, that's all.
		return nil
	}

	if mode(rate, amount, s.Pair.MinValue) == buyMode {
		return nil
	}

	// if active sell order is present, cancel it
	// before placing a new one.
	if s.cache.unconfirmedExists() {
		// retrieve cached unconfirmed order.
		unconf := s.cache.getUnconfirmed()
		// retrieve order from exchange.
		ord, err := s.Exchange.GetOrder(s.Pair, unconf.id)
		if err != nil {
			if exchErr, ok := err.(exchange.Error); ok {
				// if order does not exist in the exchange, clear it from the cache.
				if exchErr.Code == 404 {
					s.cache.cancelUnconfirmed()
				}
			} else {
				return err
			}
		} else {
			// cancel the order.
			if err := s.Exchange.CancelOrder(s.Pair, ord.ID); err != nil {
				return s.prepError(err)
			}

			// clear unconfirmed order cache.
			s.cache.cancelUnconfirmed()
		}
	}

	// place sell order
	_, err = s.Exchange.Sell(s.Pair, rate, amount)
	if err != nil {
		return s.prepError(err)
	}

	return nil
}
