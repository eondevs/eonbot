package inner

import (
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

func (i *Internal) configsRoutes() http.Handler {
	router := chi.NewRouter()

	router.Get("/summary", i.summary)

	router.Route("/main", func(r chi.Router) {
		r.Put("/", i.mainUpload)
		r.Get("/", i.mainDownload)
	})

	router.Route("/remote", func(r chi.Router) {
		r.Put("/", i.remoteUpload)
		r.Get("/", i.remoteDownload)
	})

	router.Route("/sub", func(r chi.Router) {
		r.Put("/", i.subUpload)
		r.Get("/", i.subDownload)
		r.Delete("/", i.subRemove)
	})

	router.Route("/strategy", func(r chi.Router) {
		r.Put("/", i.strategyUpload)
		r.Get("/", i.strategyDownload)
		r.Delete("/", i.strategyRemove)
	})

	return router
}

func (i *Internal) summary(w http.ResponseWriter, r *http.Request) {
	successfulJSONResp(w, i.bot.conf.Summary(), http.StatusOK)
}

/*
   main config
*/

func (i *Internal) mainUpload(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	if err := i.bot.conf.MainConfig().SetAndSaveJSON(body); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

func (i *Internal) mainDownload(w http.ResponseWriter, r *http.Request) {
	c, err := i.bot.conf.MainConfig().Download()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	jsonType(w)
	successfulResp(w, c, http.StatusOK)
}

/*
   remote config
*/

func (i *Internal) remoteUpload(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	if err := i.bot.conf.RemoteConfig().SetAndSaveJSON(body); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

func (i *Internal) remoteDownload(w http.ResponseWriter, r *http.Request) {
	c, err := i.bot.conf.RemoteConfig().Download()
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	jsonType(w)
	successfulResp(w, c, http.StatusOK)
}

/*
   sub configs
*/

func (i *Internal) subUpload(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	if err := i.bot.conf.SubConfigs().SetAndSaveJSON(body); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

func (i *Internal) subDownload(w http.ResponseWriter, r *http.Request) {
	var query struct {
		FileName string `schema:"fileName"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	c, err := i.bot.conf.SubConfigs().Download(query.FileName)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	jsonType(w)
	successfulResp(w, c, http.StatusOK)
}

func (i *Internal) subRemove(w http.ResponseWriter, r *http.Request) {
	var query struct {
		FileName string `schema:"fileName"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	if err := i.bot.conf.SubConfigs().Remove(query.FileName); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

/*
   strategies
*/

func (i *Internal) strategyUpload(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	if err := i.bot.conf.Strategies().SetAndSaveJSON(body); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}

func (i *Internal) strategyDownload(w http.ResponseWriter, r *http.Request) {
	var query struct {
		FileName string `schema:"fileName"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	c, err := i.bot.conf.Strategies().Download(query.FileName)
	if err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	jsonType(w)
	successfulResp(w, c, http.StatusOK)
}

func (i *Internal) strategyRemove(w http.ResponseWriter, r *http.Request) {
	var query struct {
		FileName string `schema:"fileName"`
	}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		reqMalformed(w)
		return
	}

	if err := i.bot.conf.Strategies().Remove(query.FileName); err != nil {
		errorResp(w, err, http.StatusBadRequest)
		return
	}

	successfulEmptyResp(w, http.StatusOK)
}
