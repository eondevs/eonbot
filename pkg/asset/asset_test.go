package asset

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/gorilla/schema"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, Asset("BTC"), New("btc   "))
}

func TestNewPair(t *testing.T) {
	assert.Equal(t, Pair{Base: "ETH", Counter: "BTC"}, NewPair("ETH", "BTC"))
}

func TestPairFromString(t *testing.T) {
	pair, err := PairFromString("ETH_BTC")
	assert.Nil(t, err)
	assert.Equal(t, Pair{Base: "ETH", Counter: "BTC"}, pair)

	_, err = PairFromString("ETH_")
	assert.NotNil(t, err)

	_, err = PairFromString("ETH")
	assert.NotNil(t, err)
}

func TestFullPairFromString(t *testing.T) {
	meta := PairMeta{
		BasePrecision: 1,
		MinValue:      decimal.RequireFromString("0.01"),
	}

	pair, err := FullPairFromString("ETH_BTC", meta)
	assert.Nil(t, err)
	assert.Equal(t, Pair{Base: "ETH", Counter: "BTC", PairMeta: meta}, pair)

	_, err = FullPairFromString("ETH_", meta)
	assert.NotNil(t, err)
}

func TestPair_IsValid(t *testing.T) {
	pair := Pair{Base: "ETH", Counter: "BTC"}
	assert.Equal(t, true, pair.IsValid())
}

func TestPair_RequireValid(t *testing.T) {
	pair := Pair{Base: "ETH"}
	assert.Equal(t, ErrPairInvalid, pair.RequireValid())

	pair = Pair{Base: "ETH", Counter: "BTC"}
	err := pair.RequireValid()
	assert.Nil(t, err)
}

func TestPair_String(t *testing.T) {
	pair := Pair{Base: "ETH", Counter: "BTC"}
	assert.Equal(t, "ETH_BTC", pair.String())

	pair = Pair{Base: "ETH"}
	assert.Equal(t, "", pair.String())
}

func TestPair_GetSharedCode(t *testing.T) {
	pair := Pair{Base: "ETH", Counter: "BTC"}
	assert.Equal(t, "ETH/BTC", pair.GetSharedCode("/", false))
	assert.Equal(t, "BTC&ETH", pair.GetSharedCode("&", true))
}

func TestPair_Equal(t *testing.T) {
	pair1 := Pair{Base: "ETH", Counter: "BTC"}
	pair2 := Pair{Base: "ETH", Counter: "USDT"}
	assert.Equal(t, false, pair1.Equal(pair2))
}

func TestPair_Prepare(t *testing.T) {
	pair := Pair{
		Base:    "ETH",
		Counter: "BTC",
		PairMeta: PairMeta{
			MaxRate:    decimal.RequireFromString("100.1"),
			MinRate:    decimal.RequireFromString("10"),
			MaxAmount:  decimal.RequireFromString("50.5"),
			MinAmount:  decimal.RequireFromString("5"),
			MinValue:   decimal.RequireFromString("60"),
			RateStep:   decimal.RequireFromString("0.05"),
			AmountStep: decimal.RequireFromString("0.1"),
		},
	}

	// max rate

	rate := decimal.RequireFromString("300.1")
	amount := decimal.RequireFromString("10.1")

	rate, amount, err := pair.Transaction(rate, amount)
	assert.NotNil(t, err)
	assert.Equal(t, decimal.Zero, rate)
	assert.Equal(t, decimal.Zero, amount)

	// min rate

	rate = decimal.RequireFromString("9")

	rate, amount, err = pair.Transaction(rate, amount)
	assert.NotNil(t, err)
	assert.Equal(t, decimal.Zero, rate)
	assert.Equal(t, decimal.Zero, amount)

	// max amount

	rate = decimal.RequireFromString("20.1")
	amount = decimal.RequireFromString("60.1")

	rate, amount, err = pair.Transaction(rate, amount)
	assert.NotNil(t, err)
	assert.Equal(t, decimal.Zero, rate)
	assert.Equal(t, decimal.Zero, amount)

	// min amount

	rate = decimal.RequireFromString("20.1")
	amount = decimal.RequireFromString("4")

	rate, amount, err = pair.Transaction(rate, amount)
	assert.NotNil(t, err)
	assert.Equal(t, decimal.Zero, rate)
	assert.Equal(t, decimal.Zero, amount)

	// min value

	rate = decimal.RequireFromString("10")
	amount = decimal.RequireFromString("5")

	rate, amount, err = pair.Transaction(rate, amount)
	assert.NotNil(t, err)
	assert.Equal(t, decimal.Zero, rate)
	assert.Equal(t, decimal.Zero, amount)

	// rounding

	rate = decimal.RequireFromString("10.12345")
	amount = decimal.RequireFromString("11.298")

	rate, amount, err = pair.Transaction(rate, amount)
	assert.Nil(t, err)
	assert.Equal(t, decimal.RequireFromString("10.1").String(), rate.String())
	assert.Equal(t, decimal.RequireFromString("11.2").String(), amount.String())
}

func TestPair_UnmarshalJSON(t *testing.T) {
	d := []byte(`{"pair": "ETH_BTC"}`)
	tmp := struct {
		Pair Pair `json:"pair"`
	}{}

	err := json.Unmarshal(d, &tmp)
	assert.Nil(t, err)
	assert.Equal(t, Pair{Base: "ETH", Counter: "BTC"}, tmp.Pair)

	d = []byte(`{"pair": "ETH_"}`)
	err = json.Unmarshal(d, &tmp)
	assert.NotNil(t, err)
}

func TestPair_MarshalJSON(t *testing.T) {
	tmp := struct {
		Pair Pair `json:"pair"`
	}{}

	tmp.Pair = Pair{Base: "ETH", Counter: "BTC"}

	d, err := json.Marshal(&tmp)

	assert.Nil(t, err)
	assert.Equal(t, `{"pair":"ETH_BTC"}`, string(d))
}

func TestPair_UnmarshalText(t *testing.T) {
	v := url.Values{}
	v.Set("pair", "ETH_BTC")
	tmp := struct {
		Pair Pair `schema:"pair"`
	}{}

	decoder := schema.NewDecoder()

	err := decoder.Decode(&tmp, v)
	assert.Nil(t, err)
	assert.Equal(t, Pair{Base: "ETH", Counter: "BTC"}, tmp.Pair)
}
