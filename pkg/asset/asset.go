package asset

import (
	"encoding/json"
	"eonbot/pkg/utils"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	ErrPairInvalid = errors.New("pair is invalid")
)

// Asset is an entity used to trade in exchange.
// Must be an upper cased string.
type Asset string

// New creates new Asset from provided string.
// Must be always used when creating new asset.
func New(s string) Asset {
	return Asset(strings.ToUpper(strings.TrimSpace(s)))
}

// Pair represents a specific market, combination of two assets
// in the exchange.
// By default, pair code consists of BASE and only then
// COUNTER asset (e.g. BASE_COUNTER).
type Pair struct {
	// Base represents base asset of the pair in the
	// exchange (holds it's string code).
	// It's the asset that's being bought/sold or shown
	// as the amount value.
	Base Asset `json:"base"`

	// Counter represents counter asset of the pair in the
	// exchange (holds it's string code).
	// It's the asset that's being paid/received after order or
	// shown as the rate per unit (of base asset) value.
	Counter Asset `json:"counter"`

	PairMeta
}

// PairMeta holds pair metadata retrieved from exchange driver.
type PairMeta struct {
	// BasePrecision defines the size of fractional-part of any number,
	// where base asset is used.
	BasePrecision int `json:"basePrecision"`

	// CounterPrecision defines the size of the fractional-part of any number,
	// where counter asset is used.
	CounterPrecision int `json:"counterPrecision"`

	// MinValue defines the minimum order value (rate * amount)
	// of this pair.
	MinValue decimal.Decimal `json:"minValue"`

	// MinRate defines the lowest allowed rate.
	MinRate decimal.Decimal `json:"minRate"`

	// MaxRate defines the highest allowed rate.
	MaxRate decimal.Decimal `json:"maxRate"`

	// Rate field incrementation step size.
	RateStep decimal.Decimal `json:"rateStep"`

	// MinAmount defines lowest allowed amount.
	MinAmount decimal.Decimal `json:"minAmount"`

	// MaxAmount defines highest allowed amount.
	MaxAmount decimal.Decimal `json:"maxAmount"`

	// Amount field incrementation step size.
	AmountStep decimal.Decimal `json:"amountStep"`
}

// NewPair creates new asset pair with specified
// base and counter assets.
func NewPair(base, counter Asset) Pair {
	return Pair{
		Base:    base,
		Counter: counter,
	}
}

// PairFromString creates new asset pair from
// string. Required format: BASE_COUNTER.
func PairFromString(s string) (Pair, error) {
	s = strings.TrimSpace(s)
	spl := strings.Split(s, "_")
	if len(spl) < 2 || spl[0] == "" || spl[1] == "" {
		return Pair{}, errors.New("pair format is invalid, correct format: BASE_COUNTER")
	}

	return NewPair(New(spl[0]), New(spl[1])), nil
}

// FullPairFromString creates new asset pair from
// string and adds additional meta data. Required format: BASE_COUNTER.
func FullPairFromString(s string, meta PairMeta) (Pair, error) {
	pair, err := PairFromString(s)
	if err != nil {
		return Pair{}, err
	}

	pair.PairMeta = meta
	return pair, nil
}

// IsValid checks if pair is usable.
func (p Pair) IsValid() bool {
	return p.Base != "" && p.Counter != ""
}

// RequireValid checks if pair is usable,
// if not, returns an error.
func (p Pair) RequireValid() error {
	if !p.IsValid() {
		return ErrPairInvalid
	}
	return nil
}

// String returns string representation
// of the asset pair, format: BASE_COUNTER.
func (p Pair) String() string {
	return p.GetSharedCode("_", false)
}

// GetSharedCode returns string of both asset pairs combined, separated by
// provided delimiter.
func (p Pair) GetSharedCode(delim string, counterFirst bool) string {
	if p.IsValid() {
		if counterFirst {
			return string(p.Counter) + delim + string(p.Base)
		}
		return string(p.Base) + delim + string(p.Counter)
	}
	return ""
}

// Equal checks if provided asset pair is the same
// as the receiver pair.
func (p Pair) Equal(pair Pair) bool {
	return p.String() == pair.String()
}

// Transaction checks if provided rate and amount are good to be used and rounds them to a needed
// step/precision (if any).
func (p Pair) Transaction(rate, amount decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if rate.GreaterThan(p.MaxRate) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("rate (%s) exceeds max allowed exchange rate (%s)", rate.String(), p.MaxRate.String())
	}

	if rate.LessThan(p.MinRate) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("rate (%s) is less than min allowed exchange rate (%s)", rate.String(), p.MinRate.String())
	}

	if amount.GreaterThan(p.MaxAmount) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("amount (%s) exceeds max allowed exchange amount (%s)", amount.String(), p.MaxAmount.String())
	}

	if amount.LessThan(p.MinAmount) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("amount (%s) is less than min allowed exchange amount (%s)", amount.String(), p.MinAmount.String())
	}

	value := rate.Mul(amount)
	if value.LessThan(p.MinValue) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("transaction value (%s) is less than min allowed exchange value (%s)", value.String(), p.MinValue.String())
	}

	if p.RateStep.GreaterThan(decimal.Zero) {
		rate = utils.RoundByStep(rate, p.RateStep, false)
	}

	if p.AmountStep.GreaterThan(decimal.Zero) {
		amount = utils.RoundByStep(amount, p.AmountStep, true)
	}

	return rate, amount, nil
}

// MarshalJSON handles asset pair conversion to JSON.
func (p *Pair) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalText handles asset pair JSON and other formats
// serialization.
func (p *Pair) UnmarshalText(d []byte) error {
	pair, err := PairFromString(string(d))
	if err != nil {
		return err
	}

	*p = pair

	return nil
}
