package inner

import (
	"encoding/json"
	"eonbot/pkg/control"
	"net/http"

	"github.com/go-chi/chi"
)

func (i *Internal) workflowRoutes() http.Handler {
	router := chi.NewRouter()
	router.Get("/state", i.state)
	router.Post("/start", i.start)
	router.Post("/stop", i.stop)
	router.Post("/restart", i.restart)
	return router
}

func (i *Internal) state(w http.ResponseWriter, r *http.Request) {
	successfulJSONResp(w, i.bot.control.State(), http.StatusOK)
}

func (i *Internal) start(w http.ResponseWriter, r *http.Request) {
	var start control.StartInfo
	if err := json.NewDecoder(r.Body).Decode(&start); err != nil {
		jsonReqMalformed(w)
		return
	}

	res := i.bot.control.Start(start, control.CauseRC)

	jsonType(w)
	successfulResp(w, res.JSON(), http.StatusOK)
}

func (i *Internal) stop(w http.ResponseWriter, r *http.Request) {
	var stop control.StopInfo
	if err := json.NewDecoder(r.Body).Decode(&stop); err != nil {
		jsonReqMalformed(w)
		return
	}

	res := i.bot.control.Stop(stop, control.CauseRC)

	jsonType(w)
	successfulResp(w, res.JSON(), http.StatusOK)
}

func (i *Internal) restart(w http.ResponseWriter, r *http.Request) {
	var data control.Commons
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonReqMalformed(w)
		return
	}

	i.bot.control.Restart(control.StartInfo{Commons: data}, control.StopInfo{}, control.CauseRC)

	successfulEmptyResp(w, http.StatusOK)
}
