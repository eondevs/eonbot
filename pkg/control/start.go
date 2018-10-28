package control

import (
	"fmt"
	"strings"
)

const (
	StartAlreadyActive startResult = iota
	StartAlreadyInitialized
	StartActivated
)

type Starter interface {
	WaitStart() <-chan StartInfo
	Start(info StartInfo, cause cause) startResult
	IsStartPending() bool
}

type StartInfo struct {
	Commons
}

type startResult int

func (s startResult) String() string {
	switch s {
	case StartAlreadyActive:
		return "Bot is already started."
	case StartAlreadyInitialized:
		return "Bot start is already initialized."
	case StartActivated:
		return "Bot is starting..."
	}

	return ""
}

func (s startResult) JSON() []byte {
	return []byte(fmt.Sprintf(`{"action":"start", "status":%d}`, int(s)))
}

type StartedState struct {
	StateMeta
}

func newStartedState(cause cause) StartedState {
	return StartedState{
		StateMeta: newStateMeta(cause, "running"),
	}
}

func (s StartedState) GetCause() cause {
	return s.StateMeta.Cause
}

func (s StartedState) IsRunning() bool {
	return true
}

func (s StartedState) String() string {
	var b strings.Builder
	b.WriteString("Bot state: running.\n")
	b.WriteString(s.StateMeta.String())
	return b.String()
}

func (s StartedState) StringShort() string {
	return "Bot is running."
}
