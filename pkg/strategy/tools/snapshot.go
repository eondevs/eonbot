package tools

import (
	"encoding/json"
	"sync"
)

/*
   Tool snapshot
*/

type SnapshotManager struct {
	sync.RWMutex

	// snap holds tool snapshot.
	snap Snapshot
}

func (s *SnapshotManager) Set(d interface{}, condsMet bool) {
	s.Lock()
	s.snap = Snapshot{
		CondsMet: condsMet,
		Data:     d,
	}
	s.Unlock()
}

func (s *SnapshotManager) Get() (res Snapshot) {
	s.RLock()
	res = s.snap
	s.RUnlock()
	return res
}

func (s *SnapshotManager) Clear() {
	s.Lock()
	s.snap = Snapshot{}
	s.Unlock()
}

type Snapshot struct {
	CondsMet bool `json:"condsMet"`

	// Data contains tools snapshot data.
	// NOTE: cannot be pointer.
	Data interface{} `json:"data"`
}

func (s Snapshot) IsEmpty() bool {
	return !s.CondsMet && s.Data == nil
}

func (s Snapshot) Full(conf json.RawMessage) FullSnapshot {
	return FullSnapshot{
		Properties: conf,
		Snapshot:   s,
	}
}

type FullSnapshot struct {
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
	Snapshot   Snapshot        `json:"snapshot"`
}

func (f FullSnapshot) SetType(t string) FullSnapshot {
	f.Type = t
	return f
}
