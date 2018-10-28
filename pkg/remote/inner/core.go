package inner

import (
	"context"
	"eonbot/pkg/config"
	"eonbot/pkg/control"
	"eonbot/pkg/db"
	"eonbot/pkg/exchange"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

type Internal struct {
	bot struct {
		conf     config.Manager
		control  control.Controller
		db       db.Manager
		exchange exchange.Exchange
	}
	conn struct {
		http struct {
			serv   *http.Server
			secret []byte
		}
		ws struct {
			sync.RWMutex
			clients map[string]*websocket.Conn
		}
	}
}

func New(conf config.Manager, control control.Controller, db db.Manager, exchange exchange.Exchange) *Internal {
	inter := &Internal{}
	inter.bot.conf = conf
	inter.bot.control = control
	inter.bot.db = db
	inter.bot.exchange = exchange
	inter.conn.ws.clients = make(map[string]*websocket.Conn)

	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(inter.basicAuth)
		// bot data endpoints
		r.Mount("/bot", inter.botDataRoutes())

		// workflow endpoints
		r.Mount("/workflow", inter.workflowRoutes())

		// configs endpoints
		r.Mount("/configs", inter.configsRoutes())

		// exchange endpoints
		r.Mount("/exchange", inter.exchangeRoutes())

		// websockets handler
		r.HandleFunc("/ws", inter.wsHandler)
	})

	inter.conn.http.serv = &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.ExecConfig().Get().RCPort),
		Handler: router,
	}

	go func() {
		err := inter.conn.http.serv.ListenAndServe()
		if err != http.ErrServerClosed {
			logrus.WithField("action", "http requests handling").Error(err)
		}
	}()

	return inter
}

func (i *Internal) Stop() {
	if i.conn.http.serv != nil {
		i.conn.http.serv.Shutdown(context.TODO())
	}

	i.getWSClients(func(cc map[string]*websocket.Conn) {
		if len(cc) > 0 {
			for id := range cc {
				i.removeWSClient(id)
			}
		}
	})
}

func (i *Internal) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !i.canAuthorize(user, pass) {
			errorResp(w, errors.New("unauthorized - username or password is incorrect"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (i *Internal) canAuthorize(username, password string) bool {
	return i.bot.conf.RemoteConfig().Get().Internal.Username == username &&
		i.bot.conf.RemoteConfig().Get().Internal.Password == password
}
