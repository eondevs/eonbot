package inner

import (
	"eonbot/pkg"
	"eonbot/pkg/asset"
	"eonbot/pkg/exchange"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func (i *Internal) botDataRoutes() http.Handler {
	router := chi.NewRouter()

	router.Get("/version", i.botVersion)

	router.Get("/time", i.botTime)

	router.Route("/orders", func(r chi.Router) {
		r.Get("/since-start", i.botOrdersSinceStart)
		r.Get("/", i.botOrders)
		r.Get("/count", i.botOrdersCount)
		r.Get("/hours-activity", i.botActivityOnHours)
	})

	router.Route("/cycles", func(r chi.Router) {
		r.Get("/", i.cycle)
		r.Get("/ids", i.cyclesIDs)
	})

	return router
}

func (i *Internal) botVersion(w http.ResponseWriter, r *http.Request) {
	jsonType(w)
	successfulResp(w, []byte(fmt.Sprintf(`{"versionCode": "%s", "versionName":"%s"}`, pkg.VersionCode, pkg.VersionName)), http.StatusOK)
}

func (i *Internal) botTime(w http.ResponseWriter, r *http.Request) {
	jsonType(w)
	successfulResp(w, []byte(fmt.Sprintf(`{"time":"%s"}`, time.Now().UTC().Format(time.RFC3339))), http.StatusOK)
}

/*
   bot orders
*/

func (i *Internal) botOrdersSinceStart(w http.ResponseWriter, r *http.Request) {
	count := i.bot.db.InMemory().OrdersSinceStart()
	jsonType(w)
	successfulResp(w, []byte(fmt.Sprintf(`{"count":%d}`, count)), http.StatusOK)
}

func (i *Internal) botOrders(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair  asset.Pair `schema:"pair"`
		Start time.Time  `schema:"start"`
		End   time.Time  `schema:"end"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	var err error
	orders := make(map[string][]exchange.BotOrder)
	if query.Pair.IsValid() {
		orders, err = i.bot.db.Persistent().GetPairOrders(query.Pair, query.Start, query.End)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	} else {
		orders, err = i.bot.db.Persistent().GetOrders(query.Start, query.End)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	}

	successfulJSONResp(w, orders, http.StatusOK)
}

func (i *Internal) botOrdersCount(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	var err error
	var count int
	if query.Pair.IsValid() {
		count, err = i.bot.db.Persistent().GetPairOrdersCount(query.Pair)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	} else {
		count, err = i.bot.db.Persistent().GetOrdersCount()
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	}

	jsonType(w)
	successfulResp(w, []byte(fmt.Sprintf(`{"count":%d}`, count)), http.StatusOK)
}

func (i *Internal) botActivityOnHours(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
		End  time.Time  `schema:"end"`
		Days int        `schema:"days"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	var err error
	counts := make([]int, 24)
	if query.Pair.IsValid() {
		counts, err = i.bot.db.Persistent().GetPairActivityOnHours(query.Pair, query.End, query.Days)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	} else {
		counts, err = i.bot.db.Persistent().GetActivityOnHours(query.End, query.Days)
		if err != nil {
			errorResp(w, err, http.StatusBadRequest)
			return
		}
	}

	successfulJSONResp(w, counts, http.StatusOK)
}

/*
   cycles
*/

func (i *Internal) cycle(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
		ID   int64      `schema:"id"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}
	cyc, err := i.bot.db.Persistent().GetPairCycle(query.Pair, query.ID)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}
	successfulJSONResp(w, cyc, http.StatusOK)
}

func (i *Internal) cyclesIDs(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Pair asset.Pair `schema:"pair"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	ids, err := i.bot.db.Persistent().GetPairCyclesIDs(query.Pair)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulJSONResp(w, ids, http.StatusOK)
}
