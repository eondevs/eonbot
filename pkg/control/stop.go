package control

import (
	"fmt"
	"strings"
)

const (
	StopAlreadyActive stopResult = iota
	StopAlreadyInitialized
	StopActivated
)

type Stopper interface {
	WaitStop() <-chan StopInfo
	Stop(info StopInfo, cause cause) stopResult
	IsStopPending() bool
}

type StopInfo struct {
	Commons
	Kill         bool `json:"kill"`
	PreventReset bool `json:"-"`
}

type stopResult int

func (s stopResult) String() string {
	switch s {
	case StopAlreadyActive:
		return "Bot is already stopped."
	case StopAlreadyInitialized:
		return "Bot stop is already initialized."
	case StopActivated:
		return "Bot is stopping..."
	}
	return ""
}

func (s stopResult) JSON() []byte {
	return []byte(fmt.Sprintf(`{"action":"stop", "status":%d}`, int(s)))
}

type StoppedState struct {
	StateMeta
}

func newStoppedState(cause cause) StoppedState {
	return StoppedState{
		StateMeta: newStateMeta(cause, "idle"),
	}
}

func (s StoppedState) GetCause() cause {
	return s.StateMeta.Cause
}

func (s StoppedState) IsRunning() bool {
	return false
}

func (s StoppedState) String() string {
	var b strings.Builder
	b.WriteString("Bot state: idle.\n")
	b.WriteString(s.StateMeta.String())
	return b.String()
}

func (s StoppedState) StringShort() string {
	return "Bot is idle."
}
