package db

import (
	"sync"
)

// InMemoryStorer defines methods
// used to store and retrieve
// specific data from in-memory
// database.
type InMomoryStorer interface {
	// Reset clears all in-memory data.
	Reset()

	// OrdersSinceStart returns total
	// count of orders since the
	// last start command.
	OrdersSinceStart() int

	// IncrOrdersSinceStart increments
	// orders since start in-memory value
	// by one.
	IncrOrdersSinceStart()
}

// memoryStore contains data
// used to save in-memory.
type memoryStore struct {
	ordMu            sync.RWMutex
	ordersSinceStart int
}

// newMemory creates new memoryStore.
func newMemory() *memoryStore {
	return &memoryStore{}
}

func (m *memoryStore) Reset() {
	m.ordMu.Lock()
	m.ordersSinceStart = 0
	m.ordMu.Unlock()
}

func (m *memoryStore) OrdersSinceStart() (count int) {
	m.ordMu.RLock()
	count = m.ordersSinceStart
	m.ordMu.RUnlock()
	return count
}

func (m *memoryStore) IncrOrdersSinceStart() {
	m.ordMu.Lock()
	m.ordersSinceStart++
	m.ordMu.Unlock()
}
