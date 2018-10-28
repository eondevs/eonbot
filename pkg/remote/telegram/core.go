package telegram

import (
	"eonbot/pkg/config"
	"eonbot/pkg/control"
	"eonbot/pkg/db"
	"errors"
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type Telegram struct {
	bot struct {
		conf    config.Manager
		control control.Controller
		db      db.Manager
	}
	conn struct {
		token       string
		client      *tgbotapi.BotAPI
		stop        chan struct{}
		subsMU      sync.RWMutex
		subscribers map[int64]struct{}
	}
}

func New(conf config.Manager, control control.Controller, db db.Manager) *Telegram {
	client, err := tgbotapi.NewBotAPI(conf.RemoteConfig().Get().Telegram.Token)
	if err != nil {
		logrus.WithField("action", "telegram bot init").Error(err)
		return nil
	}

	telegram := &Telegram{}
	telegram.bot.conf = conf
	telegram.bot.control = control
	telegram.bot.db = db
	telegram.conn.token = conf.RemoteConfig().Get().Telegram.Token
	telegram.conn.client = client
	telegram.conn.stop = make(chan struct{})
	telegram.conn.subscribers = make(map[int64]struct{})
	telegram.prepSubs()

	go func() {
		err := telegram.listen()
		if err != nil {
			logrus.WithField("action", "telegram commands handling").Error(err)
		}
	}()

	return telegram
}

func (t *Telegram) prepSubs() {
	subs, err := t.bot.db.Persistent().GetTelegramSubscribers()
	if err != nil {
		logrus.WithField("action", "telegram subscribers retrieval from db").Error(err)
		return
	}

	for _, sub := range subs {
		t.setSubscriber(sub, false)
	}
}

func (t *Telegram) listen() error {
	if t.conn.client == nil || t.conn.stop == nil {
		return errors.New("telegram client not initialized")
	}

	t.Publish("Telegram module started.")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := t.conn.client.GetUpdatesChan(u)
	if err != nil {
		return err
	}

Outer:
	for {
		select {
		case <-t.conn.stop:
			t.cleanUp()
			break Outer
		case up := <-updates:
			if up.Message == nil {
				continue
			}

			if up.Message.From.UserName != t.bot.conf.RemoteConfig().Get().Telegram.Owner {
				t.sendAndAbsorb("User not authorized.", up.Message.Chat.ID)
			}

			t.parseCMD(up.Message)
		}
	}

	return nil
}

func (t *Telegram) send(msg string, id int64) error {
	if t.conn.client == nil {
		return errors.New("telegram client not initialized")
	}

	tMsg := tgbotapi.NewMessage(id, msg)

	_, err := t.conn.client.Send(tMsg)
	if err != nil {
		return err
	}

	return nil
}

func (t *Telegram) sendAndAbsorb(msg string, id int64) {
	if err := t.send(msg, id); err != nil {
		logrus.WithField("action", "telegram message sending").Error(err)
	}
}

func (t *Telegram) IsTokenModified() bool {
	return t.conn.token != t.bot.conf.RemoteConfig().Get().Telegram.Token
}

// Publish sends message to all subscribers.
func (t *Telegram) Publish(msg string) {
	subs := t.getSubscribers()
	if subs == nil || len(subs) <= 0 {
		return
	}

	for id := range subs {
		t.sendAndAbsorb(msg, id)
	}
}

func (t *Telegram) Stop() {
	close(t.conn.stop)
}

/*
   subscriptions handling
*/

func (t *Telegram) setSubscriber(id int64, save bool) {
	t.conn.subsMU.Lock()
	t.conn.subscribers[id] = struct{}{}
	t.conn.subsMU.Unlock()
	if save {
		if err := t.bot.db.Persistent().SaveTelegramSubscriber(id); err != nil {
			logrus.WithField("action", "telegram subscriber saving to db").Error(err)
		}
	}
}

func (t *Telegram) removeSubscriber(id int64) {
	t.conn.subsMU.Lock()
	delete(t.conn.subscribers, id)
	t.conn.subsMU.Unlock()
	if err := t.bot.db.Persistent().DeleteTelegramSubscriber(id); err != nil {
		logrus.WithField("action", "telegram subscriber removal from db").Error(err)
	}
}

func (t *Telegram) subscriberExists(id int64) bool {
	t.conn.subsMU.RLock()
	_, ok := t.conn.subscribers[id]
	t.conn.subsMU.RUnlock()
	return ok
}

func (t *Telegram) getSubscribers() (subs map[int64]struct{}) {
	t.conn.subsMU.RLock()
	subs = t.conn.subscribers
	t.conn.subsMU.RUnlock()
	return subs
}

func (t *Telegram) cleanUp() {
	t.conn.subsMU.Lock()
	t.conn.subscribers = make(map[int64]struct{})
	t.conn.subsMU.Unlock()
}
