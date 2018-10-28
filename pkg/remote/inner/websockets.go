package inner

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func (i *Internal) addWSClient(c *websocket.Conn) string {
	id := uniuri.NewLen(20) + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	i.conn.ws.Lock()
	i.conn.ws.clients[id] = c
	i.conn.ws.Unlock()
	return id
}

func (i *Internal) removeWSClient(id string) {
	i.conn.ws.Lock()
	if _, ok := i.conn.ws.clients[id]; ok {
		i.conn.ws.clients[id].Close() // pointer
	}
	delete(i.conn.ws.clients, id)
	i.conn.ws.Unlock()
}

func (i *Internal) getWSClients(f func(cc map[string]*websocket.Conn)) {
	i.conn.ws.RLock()
	f(i.conn.ws.clients)
	i.conn.ws.RUnlock()
}

func (i *Internal) getWSClient(id string) (c *websocket.Conn) {
	i.conn.ws.RLock()
	c = i.conn.ws.clients[id]
	i.conn.ws.RUnlock()
	return c
}

var upgrader = websocket.Upgrader{}

func (i *Internal) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithField("action", "websocket connection upgrading").Error(err)
		return
	}

	id := i.addWSClient(conn)

	for {
		_, _, err := conn.NextReader()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logrus.WithField("action", "websocket data reading").Error(err)
			}
			i.removeWSClient(id)
			return
		}
	}
}

const (
	StateChangeEvent        = "state-update"
	PairCycleEndEvent       = "pair-cycle-end"
	CooldownActivationEvent = "cooldown-activation"
)

func (i *Internal) PublishJSON(event string) {
	i.getWSClients(func(cc map[string]*websocket.Conn) {
		for id, c := range cc {
			err := c.WriteJSON(fmt.Sprintf(`{"event":"%s"}`, event))
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logrus.WithField("action", "websocket data writing").Error(err)
				}
				i.removeWSClient(id)
			}
		}
	})
}
