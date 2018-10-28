package stream

import (
	"eonbot/pkg/asset"
	"eonbot/pkg/db"
	"eonbot/pkg/exchange"
	"eonbot/pkg/remote"
	"eonbot/pkg/settings"
	"eonbot/pkg/strategy"
	"fmt"

	"github.com/shopspring/decimal"
)

// Stream contains specific asset
// pair trading and watching data
// that persists between multiple cycles.
type Stream struct {
	// Pair specifies asset pair that
	// 'owns' this stream.
	Pair asset.Pair

	// Conf specifies stream configuration
	// data.
	Conf StreamConfig

	// RC specifies remote controller
	// object that should be used
	// for notifications, etc.
	RC remote.Manager

	// DB specifies database management
	// object that should be used to store
	// data to persistent store.
	DB db.Manager

	// Exchange specifies exchange driver
	// client that will be used to gather market
	// data and place orders.
	Exchange exchange.Exchange

	// strategies specifies strategies used
	// by this pair/stream. Pointers to strategies
	// are being used because data that will change
	// when various strategies methods are being called
	// needs to persist between multiple cycles.
	strategies []*strategy.Strategy

	// cache specifies misc stream data that should
	// persist between cycles.
	cache *cache
}

// StreamConfig contains all settings
// used by the asset pair stream.
type StreamConfig struct {
	// Config specifies pair-specific settings.
	Config settings.Pair

	// IsMain specifies whether Config field
	// was loaded from the main config or the sub
	// config.
	IsMain bool
}

// New creates new asset pair stream.
func New(pair asset.Pair, conf StreamConfig, rc remote.Manager, db db.Manager, exchange exchange.Exchange, strategies []strategy.Strategy) (*Stream, error) {
	s := &Stream{
		Pair:     pair,
		Conf:     conf,
		RC:       rc,
		DB:       db,
		Exchange: exchange,
		cache:    newCache(),
	}

	for _, str := range strategies {
		strClone, err := str.Clone()
		if err != nil {
			return nil, err
		}
		s.strategies = append(s.strategies, strClone)
	}

	return s, nil
}

// StrategiesInUse returns a map of strategies names
// and their indexes in the Stream's strategies map.
// key: strategy name; value: strategy index.
func (s *Stream) StrategiesInUse() map[string]int {
	strats := make(map[string]int)
	for i, str := range s.strategies {
		strats[str.Name()] = i
	}

	return strats
}

// UpdateStrategy updates strategy at a given index in the slice.
func (s *Stream) UpdateStrategy(index int, strategy strategy.Strategy) {
	if index < 0 || index > len(s.strategies) {
		return
	}
	s.strategies[index] = &strategy
}

// prepError decorates provided error with pair's code.
func (s *Stream) prepError(err error) error {
	return fmt.Errorf("%s execution: %s", s.Pair.String(), err.Error())
}

const (
	buyMode  = "buy"
	sellMode = "sell"
)

// BalancesPair contains balances of both counter
// and base assets of the pair.
type BalancesPair struct {
	Counter decimal.Decimal
	Base    decimal.Decimal
}

// mode returns whether it's a buy mode (base value is below min allowed value)
// or sell mode (base value is above min allowed value).
func mode(rate, amount, minVal decimal.Decimal) string {
	if amount.Mul(rate).GreaterThan(minVal) {
		return sellMode
	}
	return buyMode
}
