package outcome

import (
	"errors"
	"sync"
)

type DCA struct {
	Repeat int `json:"repeat"`
	Buy

	stateMu    sync.RWMutex
	StateIndex int
}

func (d DCA) Validate() error {
	if d.Repeat < 1 {
		return errors.New("repeat settings value cannot be zero or less")
	}

	return d.Buy.Validate()
}

func (d *DCA) Reset() {
	d.stateMu.Lock()
	d.StateIndex = 0
	d.stateMu.Unlock()
}

func (d *DCA) CanAct() (can bool) {
	d.stateMu.RLock()
	can = d.StateIndex < d.Repeat
	d.stateMu.RUnlock()
	return can
}

func (d *DCA) Increment() {
	d.stateMu.Lock()
	d.StateIndex++
	d.stateMu.Unlock()
}
