package stream

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/outcome"
	"errors"
	"fmt"
)

// activateOutcome determines the type of the outcome and calls its handler.
func (s *Stream) activateOutcome(out *outcome.Outcome, ticker exchange.TickerData, bal BalancesPair, strategy string) error {
	switch outConf := out.Conf.(type) {
	case *outcome.Buy:
		return s.buyOutcome(outConf, ticker, bal, strategy)
	case *outcome.Sell:
		return s.sellOutcome(outConf, ticker, bal, strategy)
	case *outcome.DCA:
		return s.dcaOutcome(outConf, ticker, bal, strategy)
	case *outcome.Telegram:
		return s.telegramOutcome(outConf, ticker, bal)
	case *outcome.Sandbox:
		return s.sandboxOutcome(outConf, ticker, bal)
	default:
		return errors.New("outcome type is invalid")
	}
}

// buyOutcome places buy order with specified amount and rate retrieved from ticker.
func (s *Stream) buyOutcome(buy *outcome.Buy, ticker exchange.TickerData, bal BalancesPair, strategy string) error {
	// check if another order is placed by the bot or not.
	if s.cache.unconfirmedExists() {
		return fmt.Errorf("order collision! %s strategy's order won't be placed, because another order is already opened by the bot and is not filled yet", strategy)
	}

	// retrieve specific ticker price to use as rate.
	rate := ticker.Price(buy.Price)

	// prepare amount by applying calculations with settings specified in outcome config.
	amount, err := buy.PrepAmount(bal.Base, bal.Counter, rate)
	if err != nil {
		return err
	}

	// make final checks and apply final changes (number steps, precision rounding).
	rate, amount, err = s.Pair.Transaction(rate, amount)
	if err != nil {
		return err
	}

	// place buy order.
	id, err := s.Exchange.Buy(s.Pair, rate, amount)
	if err != nil {
		return err
	}

	// cache order for later use.
	s.cache.setUnconfirmed(id, exchange.OrderSideBuy, strategy, nil)

	return nil
}

// sellOutcome places sell order with amount set to base asset balance and rate retrieved from ticker.
func (s *Stream) sellOutcome(sell *outcome.Sell, ticker exchange.TickerData, bal BalancesPair, strategy string) error {
	// check if another order is placed by the bot or not.
	if s.cache.unconfirmedExists() {
		return fmt.Errorf("order collision! %s strategy's order won't be placed, because another order is already opened by the bot and is not filled yet", strategy)
	}

	// retrieve specific ticker price to use as rate.
	rate := ticker.Price(sell.Price)

	// make final checks and apply final changes (number steps, precision rounding).
	rate, amount, err := s.Pair.Transaction(rate, bal.Base)
	if err != nil {
		return err
	}

	// place sell order.
	id, err := s.Exchange.Sell(s.Pair, rate, amount)
	if err != nil {
		return err
	}

	// cache order for later use.
	s.cache.setUnconfirmed(id, exchange.OrderSideSell, strategy, nil)

	return nil
}

// dcaOutcome places buy order with specified amount and rate retrieved from ticker.
func (s *Stream) dcaOutcome(dca *outcome.DCA, ticker exchange.TickerData, bal BalancesPair, strategy string) error {
	// check if another order is placed by the bot or not.
	if s.cache.unconfirmedExists() {
		return fmt.Errorf("order collision! %s strategy's order won't be placed, because another order is already opened by the bot and is not filled yet", strategy)
	}

	// check if another dca order is possible.
	if !dca.CanAct() {
		return nil
	}

	// retrieve specific ticker price to use as rate.
	rate := ticker.Price(dca.Price)

	// prepare amount by applying calculations with settings specified in outcome config.
	amount, err := dca.PrepAmount(bal.Base, bal.Counter, rate)
	if err != nil {
		return err
	}

	// make final checks and apply final changes (number steps, precision rounding).
	rate, amount, err = s.Pair.Transaction(rate, amount)
	if err != nil {
		return err
	}

	// place buy order.
	id, err := s.Exchange.Buy(s.Pair, rate, amount)
	if err != nil {
		return err
	}

	// cache order for later use.
	s.cache.setUnconfirmed(id, exchange.OrderSideBuy, strategy, func() {
		// when order is confirmed increment dca orders count.
		dca.Increment()
	})

	return nil
}

// telegramOutcome publishes message to telegram.
func (s *Stream) telegramOutcome(tg *outcome.Telegram, ticker exchange.TickerData, bal BalancesPair) error {
	s.RC.TelegramSend(tg.Msg())
	return nil
}

// sandboxOutcome does nothing and can be used as a placeholder or for testing purposes.
func (s *Stream) sandboxOutcome(sandbox *outcome.Sandbox, ticker exchange.TickerData, bal BalancesPair) error {
	return nil
}
