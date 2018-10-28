package inner

import (
	"encoding/json"
	"eonbot/pkg/asset"
	"eonbot/pkg/exchange"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func (i *Internal) exchangeRoutes() http.Handler {
	router := chi.NewRouter()
	router.Route("/api-info", func(r chi.Router) {
		r.Post("/", i.updateAPIInfo)
		r.Get("/", i.apiInfo)
	})
	router.Get("/cooldown-info", i.cooldown)
	router.Get("/pairs", i.pairs)
	router.Get("/intervals", i.intervals)
	router.Get("/ticker", i.ticker)
	router.Get("/candles", i.candles)
	router.Get("/balances", i.balances)
	router.Get("/open-orders", i.openOrders)
	router.Get("/order-history", i.orderHistory)
	return router
}

func (i *Internal) updateAPIInfo(w http.ResponseWriter, r *http.Request) {
	var info []exchange.APIInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		jsonReqMalformed(w)
		return
	}

	if err := i.bot.exchange.SetAPIInfo(info); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

func (i *Internal) apiInfo(w http.ResponseWriter, r *http.Request) {
	info, err := i.bot.exchange.GetAPIInfo()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, info, http.StatusOK)
}

func (i *Internal) cooldown(w http.ResponseWriter, r *http.Request) {
	cooldown, err := i.bot.exchange.GetCooldownInfo()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, cooldown, http.StatusOK)
}

func (i *Internal) pairs(w http.ResponseWriter, r *http.Request) {
	pairs, err := i.bot.exchange.GetPairs()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, pairs, http.StatusOK)
}

func (i *Internal) intervals(w http.ResponseWriter, r *http.Request) {
	intervals, err := i.bot.exchange.GetIntervals()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, intervals, http.StatusOK)
}

func (i *Internal) ticker(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	if query.Pair.IsValid() {
		tick, err := i.bot.exchange.GetTicker(query.Pair)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}

		successfulJSONResp(w, tick, http.StatusOK)
	} else {
		ticks, err := i.bot.exchange.GetTickers()
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}

		successfulJSONResp(w, ticks, http.StatusOK)
	}
}

func (i *Internal) candles(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair     asset.Pair `schema:"pair"`
		Interval int        `schema:"interval"`
		End      time.Time  `schema:"end"`
		Limit    int        `schema:"limit"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	candles, err := i.bot.exchange.GetCandles(query.Pair, query.Interval, query.End, query.Limit)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, candles, http.StatusOK)

}

func (i *Internal) balances(w http.ResponseWriter, r *http.Request) {
	balances, err := i.bot.exchange.GetBalances()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, balances, http.StatusOK)
}

func (i *Internal) openOrders(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	openOrders, err := i.bot.exchange.GetOpenOrders(query.Pair)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, openOrders, http.StatusOK)
}

func (i *Internal) orderHistory(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair  asset.Pair `schema:"pair"`
		Start time.Time  `schema:"start"`
		End   time.Time  `schema:"end"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	orderHist, err := i.bot.exchange.GetOrderHistory(query.Pair, query.Start, query.End)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, orderHist, http.StatusOK)
}
