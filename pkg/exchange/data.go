package exchange

import (
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

/*
	General
*/

// Data wraps all of the exchange
// data (Candle, TickerData, previous buy price) into one struct.
type Data struct {
	Ticker   TickerData
	Candles  []Candle
	BuyPrice decimal.Decimal
}

func NewData(tick TickerData, can []Candle) Data {
	return Data{
		Ticker:  tick,
		Candles: can,
	}
}

/*
	Candle
*/

const (
	OpenPrice  = "open"
	HighPrice  = "high"
	LowPrice   = "low"
	ClosePrice = "close"
)

var (
	ErrCandleValInvalid = errors.New("candle price is invalid")
)

// Candle contains one candle data
// returned from exchange.
type Candle struct {
	// Timestamp specifies candle time.
	Timestamp time.Time `json:"timestamp"`

	// Open specifies opening price of the candle.
	Open decimal.Decimal `json:"open"`

	// High specifies highest price of the candle.
	High decimal.Decimal `json:"high"`

	//Low specifies lowest price of the candle.
	Low decimal.Decimal `json:"low"`

	// Close specifies closing price of the candle.
	Close decimal.Decimal `json:"close"`

	// BaseVolume specifies how much of the base asset was
	// traded in the last day (in base asset).
	BaseVolume decimal.Decimal `json:"baseVolume"`

	// CounterVolume specifies how much of the base asset was
	// traded in the last day (in counter asset).
	CounterVolume decimal.Decimal `json:"counterVolume"`
}

func (c *Candle) Price(price string) decimal.Decimal {
	switch price {
	case OpenPrice:
		return c.Open
	case HighPrice:
		return c.High
	case LowPrice:
		return c.Low
	case ClosePrice:
		return c.Close
	default:
		return decimal.Zero
	}
}

func CandlePriceValid(price string) error {
	switch price {
	case OpenPrice, HighPrice, LowPrice, ClosePrice:
		return nil
	default:
		return ErrCandleValInvalid
	}
}

/*
	Ticker
*/

const (
	LastPrice           = "last"
	AskPrice            = "ask"
	BidPrice            = "bid"
	OneDayPercent       = "24hrpercent"
	TickerBaseVolume    = "basevolume"
	TickerCounterVolume = "countervolume"
)

var (
	ErrTickerPriceInvalid    = errors.New("ticker price is invalid")
	ErrTickerPropertyInvalid = errors.New("ticker property is invalid")
)

// TickerData contains ticker data (which constanly updates)
// from the exchange.
type TickerData struct {
	// LastPrice specifies last trade price.
	LastPrice decimal.Decimal `json:"lastPrice"`

	// AskPrice specifies lowest offered sell price (used to buy coin).
	AskPrice decimal.Decimal `json:"askPrice"`

	// BidPrice specifies highest offered buy price (used to sell coin).
	BidPrice decimal.Decimal `json:"bidPrice"`

	// BaseVolume specifies how much of the base asset was
	// traded in the last day (in base asset).
	BaseVolume decimal.Decimal `json:"baseVolume"`

	// CounterVolume specifies how much of the base asset was
	// traded in the last day (in counter asset).
	CounterVolume decimal.Decimal `json:"counterVolume"`

	// DayPercentChange specifies how much the price has changed
	// in the last 24 hours.
	DayPercentChange decimal.Decimal `json:"dayPercentChange"`
}

func (t *TickerData) Price(price string) decimal.Decimal {
	switch price {
	case LastPrice:
		return t.LastPrice
	case AskPrice:
		return t.AskPrice
	case BidPrice:
		return t.BidPrice
	default:
		return decimal.Zero
	}
}

func TickerPriceValid(price string) error {
	switch price {
	case LastPrice, AskPrice, BidPrice:
		return nil
	default:
		return ErrTickerPriceInvalid
	}
}

func (t *TickerData) MiscProp(prop string) decimal.Decimal {
	switch prop {
	case OneDayPercent:
		return t.DayPercentChange
	case TickerBaseVolume:
		return t.BaseVolume
	case TickerCounterVolume:
		return t.CounterVolume
	default:
		return decimal.Zero
	}
}

/*
	Order
*/

const (
	OrderSideBuy  = "buy"
	OrderSideSell = "sell"
)

// Order contains exchange order data.
// Open order doesn't have Timestamp.
type Order struct {
	// Timestamp specifies unix timestamp in seconds.
	Timestamp time.Time `json:"timestamp"`

	// ID specifies order id (format might be different on each exchange).
	ID string `json:"orderID"`

	// IsFilled specifies whether the order is completely filled.
	IsFilled bool `json:"isFilled"`

	// Amount specifies total amount of base asset.
	Amount decimal.Decimal `json:"amount"`

	// Rate specifies price of coin in counter asset.
	Rate decimal.Decimal `json:"rate"`

	// Side specifies order side (buy or sell).
	Side string `json:"side"`
}

// Total returns total order value (rate * amount).
func (o *Order) Total() decimal.Decimal {
	return o.Rate.Mul(o.Amount)
}

// BotOrder holds all order data with addition of strategy name.
type BotOrder struct {
	Order
	Strategy string `json:"strategy"`
}

func NewBotOrder(ord Order, strat string) BotOrder {
	return BotOrder{
		Order:    ord,
		Strategy: strat,
	}
}
