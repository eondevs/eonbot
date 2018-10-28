package remote

import (
	"eonbot/pkg/config"
	"eonbot/pkg/control"
	"eonbot/pkg/db"
	"eonbot/pkg/exchange"
	"eonbot/pkg/remote/inner"
	"eonbot/pkg/remote/telegram"
	"sync"
)

type Manager interface {
	ConfigTelegram()
	TelegramSend(msg string)
	InternalSend(event string)
	Stop()
}

type rc struct {
	bot struct {
		conf    config.Manager
		control control.Controller
		db      db.Manager
	}
	conn struct {
		teleMu   sync.RWMutex
		telegram *telegram.Telegram

		interMu  sync.RWMutex
		internal *inner.Internal
	}
}

func New(conf config.Manager, control control.Controller, db db.Manager, exchange exchange.Exchange) *rc {
	rc := &rc{}
	rc.bot.conf = conf
	rc.bot.control = control
	rc.bot.db = db
	rc.conn.internal = inner.New(conf, control, db, exchange)
	rc.ConfigTelegram()
	return rc
}

func (r *rc) ConfigTelegram() {
	if r.bot.conf.RemoteConfig().Get().Telegram.Enable {
		r.conn.teleMu.Lock()
		if r.conn.telegram == nil {
			r.conn.telegram = telegram.New(r.bot.conf, r.bot.control, r.bot.db)
		} else if r.conn.telegram.IsTokenModified() {
			r.stopTelegram(false)
			r.conn.telegram = telegram.New(r.bot.conf, r.bot.control, r.bot.db)
		}
		r.conn.teleMu.Unlock()
	} else {
		// channel is used internally for stopping and we don't want to block
		go r.stopTelegram(true)
	}
}

func (r *rc) TelegramSend(msg string) {
	r.conn.teleMu.RLock()
	if r.conn.telegram == nil {
		r.conn.teleMu.RUnlock()
		return
	}

	r.conn.telegram.Publish(msg)
	r.conn.teleMu.RUnlock()
}

func (r *rc) InternalSend(event string) {
	r.conn.interMu.RLock()
	if r.conn.internal == nil {
		r.conn.interMu.RUnlock()
		return
	}

	r.conn.internal.PublishJSON(event)
	r.conn.interMu.RUnlock()
}

func (r *rc) Stop() {
	r.stopTelegram(true)
	r.stopInternal()
}

func (r *rc) stopTelegram(lock bool) {
	if lock {
		r.conn.teleMu.Lock()
	}
	if r.conn.telegram != nil {
		r.conn.telegram.Stop()
		r.conn.telegram = nil
	}
	if lock {
		r.conn.teleMu.Unlock()
	}
}

func (r *rc) stopInternal() {
	r.conn.interMu.Lock()
	if r.conn.internal != nil {
		r.conn.internal.Stop()
		r.conn.internal = nil
	}
	r.conn.interMu.Unlock()
}
