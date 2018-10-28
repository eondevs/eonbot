package stream

import (
	"eonbot/pkg"
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sirupsen/logrus"
)

// Normal starts normal stream execution: handles open orders (if any),
// confirms active order (if any), gathers market and order history data
// and passes that data to the strategies.
func (s *Stream) Normal(bal BalancesPair) (pkg.Resulter, error) {
	res, err := s.handleOrders()
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res, nil
	}

	return s.handleStrategies(bal)
}

/*
	open orders
*/

// handleOrders call handleOpenOrders and confirmOrder functions.
func (s *Stream) handleOrders() (pkg.Resulter, error) {
	res, err := s.handleOpenOrders()
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res, nil
	}

	return nil, s.confirmOrder()
}

// handleOpenOrders retrieves all open orders, calculates
// their cancellation timestamp, caches it and when the specified
// time comes, the bot cancels those orders (if they were not
// filled yet).
func (s *Stream) handleOpenOrders() (pkg.Resulter, error) {
	// retrieve all open orders.
	openOrders, err := s.Exchange.GetOpenOrders(s.Pair)
	if err != nil {
		return nil, s.prepError(err)
	}

	// if no open orders exist, return.
	if openOrders == nil || len(openOrders) <= 0 {
		return nil, nil
	}

	// if cancellation is not allowed, return open orders stats.
	if !s.Conf.Config.CancelOpenOrders {
		return pkg.NewOpenOrdersResult(len(openOrders), len(openOrders), 0), nil
	}

	// clean up cached open orders.
	s.cache.cleanOpenOrders(openOrders)

	var cancelled int
	for _, ord := range openOrders {
		if ord.IsFilled {
			continue
		}

		// if open order is not cached, cache it/
		if !s.cache.openOrderExists(ord.ID) {
			s.cache.setOpenOrder(ord.ID, time.Second*time.Duration(s.Conf.Config.OpenOrderLifespan))
			continue
		}

		if !time.Now().UTC().After(s.cache.getOpenOrder(ord.ID)) {
			continue
		}

		// if current time is past open order's cancellation timestamp, cancel that
		// order.
		if err := s.Exchange.CancelOrder(s.Pair, ord.ID); err != nil {
			// if the order was cancelled, but for some reason
			// exchange kept it and returned it, we need
			// to ensure that we clear these things up.
			if exchErr, ok := err.(exchange.Error); ok {
				if exchErr.Code >= 400 {
					// retrieve order from exchange.
					ord, err := s.Exchange.GetOrder(s.Pair, ord.ID)
					if exchErr, ok := err.(exchange.Error); ok {
						// if order does not exist in the exchange, clear it from the cache.
						if exchErr.Code == 404 {
							s.cache.removeOpenOrder(ord.ID)
							continue
						}
					}

					return nil, s.prepError(err)
				}
			}

			return nil, s.prepError(err)
		}

		// after cancellation, remove open order from the cache.
		s.cache.removeOpenOrder(ord.ID)

		// increase cancelled orders count.
		cancelled++
	}

	return pkg.NewOpenOrdersResult(len(openOrders), len(openOrders)-cancelled, cancelled), nil
}

// confirmOrder checks if cached unconfirmed order is filled
// and saves it to db.
func (s *Stream) confirmOrder() error {
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
					return nil
				}
			}
			return s.prepError(err)
		}

		if ord.IsFilled {
			s.cache.confirmOrder()

			// save order to the db.
			if err := s.DB.Persistent().SavePairOrder(s.Pair, ord, unconf.strategy); err != nil {
				logrus.StandardLogger().WithField("action", "order saving to db").Error(err)
			}

			// increment order since start count.
			s.DB.InMemory().IncrOrdersSinceStart()
		} else {
			// since no more open orders are left, this is probably some error in exchange side,
			// so just cancel it.
			s.cache.cancelUnconfirmed()
		}
	}
	return nil
}

/*
	strategies
*/

// handleStrategies gathers market data and passes it to strategies checkers.
func (s *Stream) handleStrategies(bal BalancesPair) (pkg.Resulter, error) {
	// retrieve ticker from exchange.
	ticker, err := s.Exchange.GetTicker(s.Pair)
	if err != nil {
		return nil, s.prepError(err)
	}

	// get candles count.
	count := s.candlesCount(ticker, bal)

	// retrieve candles from exchange.
	candles, err := s.Exchange.GetCandles(s.Pair, s.Conf.Config.CandleInterval, time.Time{}, count)
	if err != nil {
		return nil, s.prepError(err)
	}

	// group collected data.
	data := exchange.NewData(ticker, candles)

	// if sell mode is active, retrieve and calculate buy price.
	if mode(ticker.BidPrice, bal.Base, s.Pair.MinValue) == sellMode {
		buyPrice, err := s.prepBuyPrice(bal)
		if err != nil {
			return nil, s.prepError(err)
		}

		data.BuyPrice = buyPrice
	}

	return s.act(data, bal)
}

// act gathers current active mode strategies and loops over them while passing
// market data into their checkers. If strategy returns true, its outcomes
// are activated.
func (s *Stream) act(data exchange.Data, bal BalancesPair) (pkg.Resulter, error) {
	// gather strategies by mode.
	strats := s.strategiesByMode(data.Ticker, bal)
	if strats == nil || len(strats) <= 0 {
		return nil, s.prepError(fmt.Errorf("strategies for %s mode are not specified", mode(data.Ticker.BidPrice, bal.Base, s.Pair.MinValue)))
	}

	res := pkg.NewSrategiesResult(nil)

	// loop over all strategies and
	// check if they allow to activate outcomes.
	for _, str := range strats {
		// snap takes strategy's snapshot and stores
		// it in the result map.
		snap := func() {
			res.Snapshots[str.Name()] = str.Snapshot()
		}

		// check if strategy allows to
		// activate outcome.
		ready, err := str.ReadyToAct(data)
		if err != nil {
			return nil, s.prepError(err)
		}

		// if strategy does not allow outcomes activation,
		// make its snapshot and go onto the next one.
		if !ready {
			snap()
			continue
		}

		// loop over strategy's outcomes and
		// handle every single one of them.
		for _, out := range str.Outcomes() {
			if err := s.activateOutcome(out, data.Ticker, bal, str.Name()); err != nil {
				return nil, s.prepError(err)
			}
		}

		// reset only strategies, not their outcomes (outcomes might have their own
		// cache e.g. dca).
		str.Reset(false)

		// make a snapshot on success
		snap()
	}

	return res, nil
}

/*
	helpers
*/

// candlesCount loops over strategies used by the stream in current mode
// and finds the max amount of candles needed.
func (s *Stream) candlesCount(ticker exchange.TickerData, bal BalancesPair) int {
	var count int
	for _, str := range s.strategiesByMode(ticker, bal) {
		if str.CandlesNeeded() > count {
			count = str.CandlesNeeded()
		}
	}

	return count
}

// strategiesByMode gathers all strategies of current active mode.
func (s *Stream) strategiesByMode(ticker exchange.TickerData, bal BalancesPair) []*strategy.Strategy {
	if mode := mode(ticker.BidPrice, bal.Base, s.Pair.MinValue); mode == buyMode {
		return s.buyModeStrategies()
	}
	return s.sellModeStrategies()
}

// buyModeStrategies returns all buy mode strategies of
// the stream.
func (s *Stream) buyModeStrategies() []*strategy.Strategy {
	if s.strategies == nil || len(s.strategies) <= 0 {
		return nil
	}

	strats := make([]*strategy.Strategy, 0)
	for _, strat := range s.strategies {
		if strat.Type() == strategy.SellModeStrat {
			continue
		}
		strats = append(strats, strat)
	}

	return strats
}

// sellModeStrategies returns all sell mode strategies of
// the stream.
func (s *Stream) sellModeStrategies() []*strategy.Strategy {
	if s.strategies == nil || len(s.strategies) <= 0 {
		return nil
	}

	strats := make([]*strategy.Strategy, 0)
	for _, strat := range s.strategies {
		if strat.Type() == strategy.BuyModeStrat {
			continue
		}
		strats = append(strats, strat)
	}

	return strats
}

// prepBuyPrice retrieves order history and averages buy price up until
// first sell order.
func (s *Stream) prepBuyPrice(bal BalancesPair) (decimal.Decimal, error) {
	// retrieve order history from exchange.
	// use user's specified day setting to determine the length of order
	// history.
	orderHist, err := s.Exchange.GetOrderHistory(s.Pair, time.Now().UTC().Add(-time.Hour*24*time.Duration(s.Conf.Config.OrderHistoryDayCount)), time.Time{})
	if err != nil {
		return decimal.Zero, err
	}

	if orderHist == nil || len(orderHist) <= 0 {
		return decimal.Zero, errors.New("buy price cannot be found in the order history")
	}

	// filter out all buy orders from the latest one to the first sell or 0 index.
	buyOrders := make([]exchange.Order, 0)     // oldest order must be the first one
	for i := len(orderHist) - 1; i >= 0; i-- { // first element in this loop is the latest order
		ord := orderHist[i]
		if ord.Side != exchange.OrderSideBuy {
			break
		}
		buyOrders = append([]exchange.Order{ord}, buyOrders...) // prepend
	}

	if len(buyOrders) <= 0 {
		return decimal.Zero, errors.New("buy price cannot be found in the order history")
	}

	// calculate average buy price.
	left := bal.Base
	var totalVal decimal.Decimal
	var totalAmount decimal.Decimal
	for i := len(buyOrders) - 1; i >= 0; i-- { // first order is the oldest one
		ord := buyOrders[i]
		amount := ord.Amount
		tmp := left.Sub(amount)
		if i == 0 { // oldest order
			if tmp.GreaterThan(decimal.Zero) {
				// if last order doesn't have enough amount for the balance left, use what's left
				// of the balance as an amount of the last order.
				amount = left
			}
			left = decimal.Zero
		} else {
			left = tmp
		}
		totalAmount = totalAmount.Add(amount)
		totalVal = totalVal.Add(amount.Mul(ord.Rate))
	}
	return totalVal.Div(totalAmount), nil
}
