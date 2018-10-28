package outcome

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	RandomSelection   = "random"
	RotatingSelection = "rotating"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Telegram struct {
	Selection string   `json:"selection" conform:"trim,lower"`
	Messages  []string `json:"messages"`

	rotMU    sync.Mutex
	rotIndex int

	randMU sync.Mutex
}

func (t Telegram) Validate() error {
	switch t.Selection {
	case RandomSelection, RotatingSelection:
		break
	default:
		return errors.New("selection type is invalid")
	}

	if t.Messages == nil || len(t.Messages) <= 0 {
		return errors.New("telegram messages list cannot be empty")
	}

	return nil
}

func (t *Telegram) Reset() {
	t.rotMU.Lock()
	t.rotIndex = 0
	t.rotMU.Unlock()
}

func (t *Telegram) Msg() string {
	switch t.Selection {
	case RandomSelection:
		return t.randMsg()
	case RotatingSelection:
		return t.rotMsg()
	default:
		return ""
	}
}

func (t *Telegram) randMsg() (msg string) {
	t.randMU.Lock()
	msg = t.Messages[rand.Intn(len(t.Messages))]
	t.randMU.Unlock()
	return msg
}

func (t *Telegram) rotMsg() (msg string) {
	t.rotMU.Lock()
	msg = t.Messages[t.rotIndex]
	t.rotIndex++
	if t.rotIndex >= len(t.Messages) {
		t.rotIndex = 0
	}
	t.rotMU.Unlock()
	return msg
}
