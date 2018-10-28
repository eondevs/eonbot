package exchange

import (
	"bytes"
	"encoding/json"
	"eonbot/pkg/asset"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrExchangeDriverAddrInvalid = NewPlainError("exchange driver address is invalid", 0)
)

type Exchange interface {
	ExchangeConn
	ExchangeConfirmer
}

// ExchangeConn specifies methods used to connect/communicate with the
// exchange driver.
type ExchangeConn interface {
	/*
		exchange client configuration
	*/

	SetAddress(addr string) error
	GetAddress() string

	/*
		exchange driver configuration
	*/

	// SetAPIInfo sends new slice of APIInfo objects to
	// the exchange driver.
	SetAPIInfo(info []APIInfo) error

	// GetAPIInfo retrieves all APIInfo objects from the
	// exchange driver.
	GetAPIInfo() ([]APIInfo, error)

	/*
		state checking
	*/

	Ping() error
	GetCooldownInfo() (CooldownInfo, error)

	/*
		exchange data gathering
	*/

	// GetIntervals retrieves allowed candles intervals from the
	// exchange driver. Return type should be integer (represents seconds).
	GetIntervals() ([]int, error)

	// GetPairs retrieves all possible pairs from the
	// exchange driver.
	GetPairs() ([]asset.Pair, error)

	// GetTicker retrieves specific pair latest ticker data from the
	// exchange driver.
	GetTicker(pair asset.Pair) (TickerData, error)

	// GetTickers retrieves tickers of all pairs.
	GetTickers() (map[string]TickerData, error)

	// GetCandles retrieves specific pair candles data from the exchange driver.
	// Interval represents seconds.
	// If latest candles data is needed, pass zero-value end parameter.
	GetCandles(pair asset.Pair, interval int, end time.Time, limit int) ([]Candle, error)

	// GetBalances retrieves balances from the exchange driver.
	GetBalances() (map[string]decimal.Decimal, error)

	// Buy places buy order via the exchange driver.
	Buy(pair asset.Pair, rate, amount decimal.Decimal) (string, error)

	// Sell places sell order via the exchange driver.
	Sell(pair asset.Pair, rate, amount decimal.Decimal) (string, error)

	// CancelOrder cancels open order via the exchange driver.
	CancelOrder(pair asset.Pair, id string) error

	// GetOrder retrieves specific order info from the exchange driver.
	GetOrder(pair asset.Pair, id string) (Order, error)

	// GetOpenOrders retrieves all open orders of the specific pair from
	// the exchange driver.
	GetOpenOrders(pair asset.Pair) ([]Order, error)

	// GetOrderHistory retrieves specific pair order history from the
	// If latest orders data is needed, pass zero-value end parameter.
	GetOrderHistory(pair asset.Pair, start, end time.Time) ([]Order, error)
}

// APIInfo holds exchange API authentication
// data.
type APIInfo struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// CooldownInfo holds exchange driver cooldown state info.
type CooldownInfo struct {
	Active bool      `json:"active"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type ExchangeConfirmer interface {
	// ConfirmPairs checks if provided pairs are
	// allowed and returns them with additional metadata.
	ConfirmPairs(pairs []asset.Pair) ([]asset.Pair, error)

	// ConfirmInterval checks if provided interval is valid.
	ConfirmInterval(interval int) error
}

// ExchangeClient is an implementation of
//
type ExchangeClient struct {
	driverAddr url.URL
	client     *http.Client
}

func New(timeout int64) *ExchangeClient {
	return &ExchangeClient{
		client: &http.Client{
			Timeout: time.Second * time.Duration(timeout),
		},
	}
}

func (e *ExchangeClient) decodeResp(resp *http.Response, target interface{}) error {
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewError(err)
	}

	if resp.StatusCode >= 400 {
		var respErr struct {
			Error string `json:"error"`
		}

		if strings.Contains(string(response), "error") {
			if err := json.Unmarshal(response, &respErr); err != nil {
				return NewError(err)
			}
		} else {
			respErr.Error = http.StatusText(resp.StatusCode)
		}

		return NewPlainError(respErr.Error, resp.StatusCode)
	} else {
		if target != nil {
			if err := json.Unmarshal(response, target); err != nil {
				return NewError(err)
			}
		}
	}

	return nil
}

func (e *ExchangeClient) SetAddress(addr string) error {
	if addr == "" {
		return ErrExchangeDriverAddrInvalid
	}

	u, err := url.Parse(addr)
	if err != nil {
		return ErrExchangeDriverAddrInvalid
	}

	e.driverAddr = *u
	return nil
}

func (e *ExchangeClient) GetAddress() string {
	return e.driverAddr.String()
}

func (e *ExchangeClient) SetAPIInfo(info []APIInfo) error {
	if e.driverAddr.String() == "" {
		return ErrExchangeDriverAddrInvalid
	}

	if len(info) == 0 {
		return errors.New("api info list cannot be empty")
	}

	jsonBody, err := json.Marshal(info)
	if err != nil {
		return NewError(err)
	}

	u := e.driverAddr
	u.Path = "api-info"

	resp, err := e.client.Post(u.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return NewError(err)
	}

	return e.decodeResp(resp, nil)
}

func (e *ExchangeClient) GetAPIInfo() ([]APIInfo, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "api-info"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	info := make([]APIInfo, 0)

	if err = e.decodeResp(resp, &info); err != nil {
		return nil, err
	}

	return info, nil
}

func (e *ExchangeClient) Ping() error {
	if e.driverAddr.String() == "" {
		return ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "ping"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return NewError(err)
	}

	return e.decodeResp(resp, nil)
}

func (e *ExchangeClient) GetCooldownInfo() (CooldownInfo, error) {
	if e.driverAddr.String() == "" {
		return CooldownInfo{}, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "cooldown-info"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return CooldownInfo{}, NewError(err)
	}

	cooldown := CooldownInfo{}

	if err = e.decodeResp(resp, &cooldown); err != nil {
		return CooldownInfo{}, err
	}

	return cooldown, nil
}

func (e *ExchangeClient) GetIntervals() ([]int, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "intervals"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	intervals := make([]int, 0)

	if err = e.decodeResp(resp, &intervals); err != nil {
		return nil, err
	}

	return intervals, nil
}

func (e *ExchangeClient) GetPairs() ([]asset.Pair, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "pairs"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	pairsMap := make(map[string]asset.PairMeta)

	if err = e.decodeResp(resp, &pairsMap); err != nil {
		return nil, err
	}

	pairs := make([]asset.Pair, 0)

	for code, data := range pairsMap {
		pair, err := asset.FullPairFromString(code, data)
		if err != nil {
			return nil, NewError(err)
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func (e *ExchangeClient) GetTicker(pair asset.Pair) (TickerData, error) {
	if e.driverAddr.String() == "" {
		return TickerData{}, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "ticker"
	q := u.Query()
	q.Set("pair", pair.String())
	u.RawQuery = q.Encode()

	resp, err := e.client.Get(u.String())
	if err != nil {
		return TickerData{}, NewError(err)
	}

	ticker := TickerData{}

	if err = e.decodeResp(resp, &ticker); err != nil {
		return TickerData{}, err
	}

	return ticker, nil
}

func (e *ExchangeClient) GetTickers() (map[string]TickerData, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "ticker"
	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	tickers := make(map[string]TickerData, 0)

	if err = e.decodeResp(resp, &tickers); err != nil {
		return nil, err
	}

	return tickers, nil
}

func (e *ExchangeClient) GetCandles(pair asset.Pair, interval int, end time.Time, limit int) ([]Candle, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return nil, NewError(err)
	}

	if err := e.ConfirmInterval(interval); err != nil {
		return nil, NewError(err)
	}

	u := e.driverAddr
	u.Path = "candles"
	q := u.Query()
	q.Set("pair", pair.String())
	q.Set("interval", fmt.Sprint(interval))
	if !end.IsZero() {
		q.Set("end", end.Format(time.RFC3339))
	}

	if limit > 0 {
		q.Set("limit", fmt.Sprint(limit))
	}

	u.RawQuery = q.Encode()

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	candles := make([]Candle, 0)

	if err = e.decodeResp(resp, &candles); err != nil {
		return nil, err
	}

	if limit > 0 {
		if len(candles) > limit {
			candles = candles[len(candles)-limit:]
		} else if len(candles) < limit {
			return nil, NewPlainError(fmt.Sprintf("returned candles list size (%d) hasn't met minimal required size (%d)", len(candles), limit), 0)
		}
	}

	return candles, nil
}

func (e *ExchangeClient) GetBalances() (map[string]decimal.Decimal, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	u := e.driverAddr
	u.Path = "balances"

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	balances := make(map[string]decimal.Decimal)
	if err = e.decodeResp(resp, &balances); err != nil {
		return nil, err
	}

	return balances, nil
}

func (e *ExchangeClient) Buy(pair asset.Pair, rate, amount decimal.Decimal) (string, error) {
	if e.driverAddr.String() == "" {
		return "", ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return "", NewError(err)
	}

	if rate.IsZero() || rate.IsNegative() {
		return "", NewPlainError("rate is invalid", 0)
	}

	if amount.IsZero() || amount.IsNegative() {
		return "", NewPlainError("amount is invalid", 0)
	}

	u := e.driverAddr
	u.Path = "buy"

	data := struct {
		Pair   string          `json:"pair"`
		Rate   decimal.Decimal `json:"rate"`
		Amount decimal.Decimal `json:"amount"`
	}{
		Pair:   pair.String(),
		Rate:   rate,
		Amount: amount,
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return "", NewError(err)
	}

	resp, err := e.client.Post(u.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", NewError(err)
	}

	var id struct {
		ID string `json:"id"`
	}

	if err = e.decodeResp(resp, &id); err != nil {
		return "", err
	}

	return id.ID, nil
}

func (e *ExchangeClient) Sell(pair asset.Pair, rate, amount decimal.Decimal) (string, error) {
	if e.driverAddr.String() == "" {
		return "", ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return "", NewError(err)
	}

	if rate.IsZero() || rate.IsNegative() {
		return "", NewPlainError("rate is invalid", 0)
	}

	if amount.IsZero() || amount.IsNegative() {
		return "", NewPlainError("amount is invalid", 0)
	}

	u := e.driverAddr
	u.Path = "sell"

	data := struct {
		Pair   string          `json:"pair"`
		Rate   decimal.Decimal `json:"rate"`
		Amount decimal.Decimal `json:"amount"`
	}{
		Pair:   pair.String(),
		Rate:   rate,
		Amount: amount,
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return "", NewError(err)
	}

	resp, err := e.client.Post(u.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", NewError(err)
	}

	var id struct {
		ID string `json:"id"`
	}

	if err = e.decodeResp(resp, &id); err != nil {
		return "", err
	}

	return id.ID, nil
}

func (e *ExchangeClient) CancelOrder(pair asset.Pair, id string) error {
	if e.driverAddr.String() == "" {
		return ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return NewError(err)
	}

	if id == "" {
		return NewPlainError("id is invalid", 0)
	}

	u := e.driverAddr
	u.Path = "cancel"
	data := struct {
		Pair string `json:"pair"`
		ID   string `json:"id"`
	}{
		Pair: pair.String(),
		ID:   id,
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return NewError(err)
	}

	resp, err := e.client.Post(u.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return NewError(err)
	}

	return e.decodeResp(resp, nil)
}

func (e *ExchangeClient) GetOrder(pair asset.Pair, id string) (Order, error) {
	if e.driverAddr.String() == "" {
		return Order{}, ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return Order{}, NewError(err)
	}

	if id == "" {
		return Order{}, NewPlainError("id is invalid", 0)
	}

	u := e.driverAddr
	u.Path = "order"
	q := u.Query()
	q.Set("pair", pair.String())
	q.Set("id", id)
	u.RawQuery = q.Encode()

	resp, err := e.client.Get(u.String())
	if err != nil {
		return Order{}, NewError(err)
	}

	order := Order{}

	if err = e.decodeResp(resp, &order); err != nil {
		return Order{}, err
	}

	return order, nil
}

func (e *ExchangeClient) GetOpenOrders(pair asset.Pair) ([]Order, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return nil, NewError(err)
	}

	u := e.driverAddr
	u.Path = "open-orders"
	q := u.Query()
	q.Set("pair", pair.String())
	u.RawQuery = q.Encode()

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	openOrders := make([]Order, 0)
	if err = e.decodeResp(resp, &openOrders); err != nil {
		return nil, err
	}

	return openOrders, nil
}

func (e *ExchangeClient) GetOrderHistory(pair asset.Pair, start, end time.Time) ([]Order, error) {
	if e.driverAddr.String() == "" {
		return nil, ErrExchangeDriverAddrInvalid
	}

	if err := pair.RequireValid(); err != nil {
		return nil, NewError(err)
	}

	u := e.driverAddr
	u.Path = "order-history"
	q := u.Query()
	q.Set("pair", pair.String())
	if !start.IsZero() {
		q.Set("start", start.Format(time.RFC3339))
	}

	if !end.IsZero() {
		q.Set("end", end.Format(time.RFC3339))
	}

	u.RawQuery = q.Encode()

	resp, err := e.client.Get(u.String())
	if err != nil {
		return nil, NewError(err)
	}

	orderHist := make([]Order, 0)
	if err = e.decodeResp(resp, &orderHist); err != nil {
		return nil, err
	}

	return orderHist, nil
}

func (e *ExchangeClient) ConfirmPairs(pairs []asset.Pair) ([]asset.Pair, error) {
	if len(pairs) <= 0 {
		return nil, errors.New("active pairs list cannot be empty")
	}

	exchPairs, err := e.GetPairs()
	if err != nil {
		return nil, err
	}

	res := make([]asset.Pair, 0)
	for _, pair := range pairs {
		var found bool
		for _, exchPair := range exchPairs {
			if exchPair.Equal(pair) {
				found = true
				res = append(res, exchPair)
			}
		}

		if !found {
			return nil, fmt.Errorf("%s pair is not allowed", pair.String())
		}
	}

	return res, nil
}

func (e *ExchangeClient) ConfirmInterval(interval int) error {
	if interval == 0 {
		return errors.New("interval is invalid")
	}

	exchIntervals, err := e.GetIntervals()
	if err != nil {
		return err
	}

	for _, exchInterval := range exchIntervals {
		if exchInterval == interval {
			return nil
		}
	}

	return fmt.Errorf("%d interval is not allowed", interval)
}

// Error defines exchange error.
type Error struct {
	Code int
	Msg  string
}

func NewPlainError(msg string, code int) Error {
	return Error{
		Code: code,
		Msg:  msg,
	}
}

func NewError(err error) Error {
	return Error{
		Msg: err.Error(),
	}
}

func (e Error) Error() string {
	var b strings.Builder
	b.WriteString("exchange")
	if e.Code != 0 {
		b.WriteString(fmt.Sprintf(" (code: %d)", e.Code))
	}
	b.WriteString(fmt.Sprintf(": %s", e.Msg))
	return b.String()
}
