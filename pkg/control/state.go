package control

import (
	"fmt"
	"strings"
	"time"
)

const (
	CauseInit cause = iota
	CauseConfigChange
	CauseConfigLoadErr
	CauseRC
	CauseInvalidEonBotAuth
)

type Stater interface {
	State() StateInfoer
	updateState(s StateInfoer)
}

type StateInfoer interface {
	GetCause() cause
	IsRunning() bool
	String() string
	StringShort() string
}

type cause int

func (c cause) String() string {
	switch c {
	case CauseInit:
		return "init"
	case CauseConfigChange:
		return "config change"
	case CauseConfigLoadErr:
		return "config loading/parsing error"
	case CauseRC:
		return "remote control"
	case CauseInvalidEonBotAuth:
		return "problems with eonbot authentication"
	default:
		return ""
	}
}

type StateMeta struct {
	// State specifies state string name.
	State string `json:"state"`

	// Cause specifies what triggered this state.
	Cause cause `json:"cause"`

	// ActivationTime specifies time when this state was entered.
	ActivationTime time.Time `json:"activationTime"`
}

func newStateMeta(cause cause, state string) StateMeta {
	return StateMeta{
		State:          state,
		Cause:          cause,
		ActivationTime: time.Now().UTC(),
	}
}

func (s *StateMeta) ActiveTime() time.Duration {
	return time.Since(s.ActivationTime)
}

func (s *StateMeta) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("State active time: %s.\n", s.ActiveTime().String()))
	b.WriteString(fmt.Sprintf("State was activated by: %s.\n", s.Cause.String()))
	return b.String()
}
