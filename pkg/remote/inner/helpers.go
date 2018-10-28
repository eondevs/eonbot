package inner

import (
	"encoding/json"
	"eonbot/pkg/db"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func jsonType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func errorResp(w http.ResponseWriter, err error, code int) {
	switch e := err.(type) {
	case exchange.Error:
		err = errors.New(e.Msg)
		if e.Code > 0 {
			code = e.Code
		}
	default:
		if err == db.ErrDataNotFound || err == file.ErrFileNotFound {
			code = http.StatusNotFound
		}
	}
	jsonType(w)
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
}

func jsonReqMalformed(w http.ResponseWriter) {
	errorResp(w, errors.New("request JSON body is malformed"), http.StatusBadRequest)
}

func reqMalformed(w http.ResponseWriter) {
	errorResp(w, errors.New("request data is malformed"), http.StatusBadRequest)
}

func successfulResp(w http.ResponseWriter, data []byte, code int) {
	jsonType(w)
	w.WriteHeader(code)
	fmt.Fprint(w, string(data))
}

func successfulJSONResp(w http.ResponseWriter, v interface{}, code int) {
	jsonType(w)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		errorResp(w, errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError)
		logrus.WithField("action", "http response marshaling").Error(err)
	}
}

func successfulEmptyResp(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
