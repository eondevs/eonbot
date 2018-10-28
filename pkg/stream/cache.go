package stream

import (
	"eonbot/pkg/exchange"
	"time"
)

// cache holds data that should persist between cycles.
// Mutexes are not needed because cache is used
// synchronously in the same stream/goroutine.
type cache struct {
	// openOrders specifies a map
	// of open orders and their calculated
	// cancellation timestamp.
	openOrders map[string]time.Time

	// unconfirmed specifies open order
	// that the bot waits to be filled.
	unconfirmed unconfirmed
}

// newCache creates new cache pointer.
func newCache() *cache {
	return &cache{
		openOrders: make(map[string]time.Time),
	}
}

/*
   Open orders cache
*/

// setOpenOrder sets open order with the calculated cancellation
// timestamp to the open orders map.
// Duration parameter is added to the current timestamp to
// create cancellation timestamp.
func (c *cache) setOpenOrder(id string, d time.Duration) {
	c.openOrders[id] = time.Now().UTC().Add(d)
}

// removeOpenOrder removes open order from the open orders map.
func (c *cache) removeOpenOrder(id string) {
	delete(c.openOrders, id)
}

// getOpenOrder returns open order by specified order id.
func (c *cache) getOpenOrder(id string) time.Time {
	return c.openOrders[id]
}

// getOpenOrders returns all open orders currently in
// open orders map.
func (c *cache) getOpenOrders() map[string]time.Time {
	return c.openOrders
}

// openOrderExists returns true if order with the
// provided id exists in the open orders map.
func (c *cache) openOrderExists(id string) bool {
	return !c.openOrders[id].IsZero()
}

// cleanOpenOrders removes those orders from the
// open orders map that do not exist in the provided
// open orders slice.
func (c *cache) cleanOpenOrders(orders []exchange.Order) {
Outer:
	for id := range c.getOpenOrders() {
		for _, o := range orders {
			if o.ID == id {
				continue Outer
			}
		}

		c.removeOpenOrder(id)
	}
}

/*
   Unconfirmed stream order
*/

// unconfirmed contains order's info that the
// bot waits to be filled.
type unconfirmed struct {
	// id specifies unconfirmed order id.
	id string

	// side specifies unconfirmed order side (buy/sell).
	side string

	// strategy specifies what strategy was used to place
	// unconfirmed order.
	strategy string

	// confirmCb specifies callback that will be called
	// when unconfirmed order will be confirmed.
	confirmCb func()
}

// unconfirmedExists specifies whether the
// unconfirmed order exists or not.
func (c *cache) unconfirmedExists() bool {
	return c.unconfirmed.id != ""
}

// setUnconfirmed sets provided unconfirmed order data to the cache.
func (c *cache) setUnconfirmed(id string, side string, strategy string, cb func()) {
	c.unconfirmed.id = id
	c.unconfirmed.side = side
	c.unconfirmed.strategy = strategy
	c.unconfirmed.confirmCb = cb
}

// getUnconfirmed returns current unconfirmed order info.
func (c *cache) getUnconfirmed() unconfirmed {
	return c.unconfirmed
}

// cancelUnconfirmed clears unconfirmed order cache info.
func (c *cache) cancelUnconfirmed() {
	c.unconfirmed.id = ""
	c.unconfirmed.strategy = ""
	c.unconfirmed.side = ""
	c.unconfirmed.confirmCb = nil
}

// confirmOrder clears unconfirmed order cache info and
// call confirmation callback.
func (c *cache) confirmOrder() {
	if c.unconfirmed.confirmCb != nil {
		c.unconfirmed.confirmCb()
	}
	c.cancelUnconfirmed()
}
