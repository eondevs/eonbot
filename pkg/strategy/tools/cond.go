package tools

import (
	"encoding/json"
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/indicators"
	"eonbot/pkg/strategy/indicators/ma"
	"errors"

	"github.com/leebenson/conform"
	"github.com/shopspring/decimal"
)

var (
	ErrNoPossibleConditions        = errors.New("no possible conditions")
	ErrCondObjectInvalid           = errors.New("condition object is invalid")
	ErrDataRetrieverNotInitialized = errors.New("data retriever not initialized")
	ErrExchValInvalid              = errors.New("exchange value is invalid")
)

const (
	CondEqual        = "equal"
	CondAbove        = "above"
	CondAboveOrEqual = "aboveorequal"
	CondBelow        = "below"
	CondBelowOrEqual = "beloworequal"
	CondAboveOrBelow = "aboveorbelow"
)

type Cond struct {
	C string `json:"cond" conform:"trim,lower"`
}

func (c *Cond) Match(val1, val2 decimal.Decimal) bool {
	switch c.C {
	case CondEqual:
		return val1.Equal(val2)
	case CondAbove:
		return val1.GreaterThan(val2)
	case CondAboveOrEqual:
		return val1.GreaterThanOrEqual(val2)
	case CondBelow:
		return val1.LessThan(val2)
	case CondBelowOrEqual:
		return val1.LessThanOrEqual(val2)
	case CondAboveOrBelow:
		return val1.LessThan(val2) || val1.GreaterThan(val2)
	default:
		return false
	}
}

func (c *Cond) Validate() error {
	switch c.C {
	case CondEqual, CondAbove, CondAboveOrEqual, CondBelow, CondBelowOrEqual, CondAboveOrBelow:
		break
	default:
		return ErrNoPossibleConditions
	}

	return nil
}

type CondObject struct {
	//config
	Obj     string          `json:"obj" conform:"trim,lower"`
	ObjConf json.RawMessage `json:"objConf"`

	// internal
	getData         func(d exchange.Data) (decimal.Decimal, error)
	getCandlesCount func() int

	tickerPrice    bool
	tickerMiscProp bool
	candlePrice    bool
	ma             bool
}

func (c *CondObject) AllowTickerPrice() {
	c.tickerPrice = true
}

func (c *CondObject) AllowTickerMiscProp() {
	c.tickerMiscProp = true
}

func (c *CondObject) AllowCandlePrice() {
	c.candlePrice = true
}

func (c *CondObject) AllowMA() {
	c.ma = true
}

func (c *CondObject) Init(offset int) error {
	if offset < 0 {
		return errors.New("offset cannot be negative")
	}

	switch c.Obj {
	case exchange.LastPrice, exchange.AskPrice, exchange.BidPrice:
		if !c.tickerPrice {
			return ErrCondObjectInvalid
		}
		c.getData = func(d exchange.Data) (decimal.Decimal, error) {
			return d.Ticker.Price(c.Obj), nil
		}

		c.getCandlesCount = func() int {
			return 0
		}
		return nil
	case exchange.OneDayPercent, exchange.TickerBaseVolume, exchange.TickerCounterVolume:
		if !c.tickerMiscProp {
			return ErrCondObjectInvalid
		}
		c.getData = func(d exchange.Data) (decimal.Decimal, error) {
			return d.Ticker.MiscProp(c.Obj), nil
		}

		c.getCandlesCount = func() int {
			return 0
		}
		return nil
	case exchange.OpenPrice, exchange.HighPrice, exchange.LowPrice, exchange.ClosePrice:
		if !c.candlePrice {
			return ErrCondObjectInvalid
		}
		c.getData = func(d exchange.Data) (decimal.Decimal, error) {
			if d.Candles == nil || len(d.Candles)-offset <= 0 {
				return decimal.Zero, errors.New("candles list size is too small")
			}

			candle := d.Candles[len(d.Candles)-offset-1]
			return candle.Price(c.Obj), nil
		}

		c.getCandlesCount = func() int {
			return offset + 1
		}
		return nil
	case ma.SMAName, ma.EMAName, ma.WMAName:
		if !c.ma {
			return ErrCondObjectInvalid
		}
		var conf ma.MAConfig
		if err := json.Unmarshal(c.ObjConf, &conf); err != nil {
			return err
		}

		conform.Strings(&conf)

		if err := conf.Validate(); err != nil {
			return err
		}

		ma, err := indicators.NewMAFromConfig(c.Obj, conf, offset)
		if err != nil {
			return err
		}

		c.getData = func(d exchange.Data) (decimal.Decimal, error) {
			return ma.Calc(d.Candles)
		}

		c.getCandlesCount = func() int {
			return ma.CandlesCount()
		}
		return nil
	default:
		return ErrCondObjectInvalid
	}
}

func (c *CondObject) Validate() error {
	switch c.Obj {
	case exchange.LastPrice, exchange.AskPrice, exchange.BidPrice:
		if !c.tickerPrice {
			return ErrCondObjectInvalid
		}
		break
	case exchange.OneDayPercent, exchange.TickerBaseVolume, exchange.TickerCounterVolume:
		if !c.tickerMiscProp {
			return ErrCondObjectInvalid
		}
		break
	case exchange.OpenPrice, exchange.HighPrice, exchange.LowPrice, exchange.ClosePrice:
		if !c.candlePrice {
			return ErrCondObjectInvalid
		}
		break
	case ma.SMAName, ma.EMAName, ma.WMAName:
		if !c.ma {
			return ErrCondObjectInvalid
		}
		break
	default:
		return ErrCondObjectInvalid
	}

	return nil
}

func (c *CondObject) Value(d exchange.Data) (decimal.Decimal, error) {
	if c.getData == nil {
		return decimal.Zero, ErrDataRetrieverNotInitialized
	}

	val, err := c.getData(d)
	if err != nil {
		return decimal.Zero, err
	}

	if val.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrExchValInvalid
	}

	return val, nil
}

func (c *CondObject) CandlesCount() int {
	if c.getCandlesCount == nil {
		return 0
	}
	return c.getCandlesCount()
}

func (c CondObject) Snapshot(val decimal.Decimal) CondObjectSnapshot {
	return NewCondObjectSnapshot(val)
}

type CondObjectSnapshot struct {
	Value decimal.Decimal `json:"objVal"`
}

func NewCondObjectSnapshot(val decimal.Decimal) CondObjectSnapshot {
	return CondObjectSnapshot{
		Value: val,
	}
}
