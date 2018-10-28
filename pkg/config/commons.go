package config

import (
	"sync"
	"time"
)

const (
	NotExists ChangeStatus = iota
	NotChanged
	Changed
)

type ChangeStatus int

type commons interface {
	// SetAndSaveJSON unmarshals provided json,
	// formats it, saves it to file and sets it in config struct.
	SetAndSaveJSON(d []byte) error

	// IsChanged checks if configer internal config values were changed.
	IsChanged() bool

	// UpdateCheckTime updates last check time.
	UpdateCheckTime()

	modTime() time.Time
	setModTime(t time.Time)

	// Load checks if file was modified and loads it.
	// If strict is true and ifconfigs don't exist,
	// error will be returned.
	Load(strict bool) error

	// IsLoaded returns true if config(s) were loaded or not.
	IsLoaded() bool
}

type singularCommons interface {
	Download() ([]byte, error)
	commons
}

type multipleCommons interface {
	Download(fileN string) ([]byte, error)
	Remove(fileN string) error
	commons
}

type confUtils struct {
	confMod
	lastChangeCheck

	exec ExecConfiger

	// for other read/writes
	sync.RWMutex
}

func (c *confUtils) IsChanged() bool {
	return c.modTime().After(c.lastCheck())
}

func (c *confUtils) UpdateCheckTime() {
	c.updateLastCheck()
}

type confMod struct {
	mMu     sync.RWMutex
	lastMod time.Time
	loaded  bool
}

func (c *confMod) modTime() (t time.Time) {
	c.mMu.RLock()
	t = c.lastMod
	c.mMu.RUnlock()
	return t
}

func (c *confMod) setModTime(t time.Time) {
	c.mMu.Lock()
	c.lastMod = t
	c.mMu.Unlock()
}

func (c *confMod) notLoaded() {
	c.mMu.Lock()
	c.loaded = false
	c.mMu.Unlock()
}

func (c *confMod) wasLoaded() {
	c.mMu.Lock()
	c.loaded = true
	c.mMu.Unlock()
}

func (c *confMod) IsLoaded() (l bool) {
	c.mMu.Lock()
	l = c.loaded
	c.mMu.Unlock()
	return l
}

type lastChangeCheck struct {
	lMu             sync.RWMutex
	lastChangeCheck time.Time
}

func (l *lastChangeCheck) lastCheck() (t time.Time) {
	l.lMu.RLock()
	t = l.lastChangeCheck
	l.lMu.RUnlock()
	return t
}

func (l *lastChangeCheck) updateLastCheck() {
	l.lMu.Lock()
	l.lastChangeCheck = time.Now().UTC()
	l.lMu.Unlock()
}

func (l *lastChangeCheck) resetLastCheck() {
	l.lMu.Lock()
	l.lastChangeCheck = time.Unix(0, 0)
	l.lMu.Unlock()
}
