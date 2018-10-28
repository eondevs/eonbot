package config

import "eonbot/pkg/settings"

type ExecConfiger interface {
	// Get returns ExecConfig that was
	// passed at the init of Configer.
	Get() settings.Exec
}

type exec struct {
	// conf is immutable config, set only at the init.
	conf settings.Exec
}

func newExec(conf settings.Exec) *exec {
	return &exec{conf: conf}
}

func (e *exec) Get() settings.Exec {
	return e.conf
}
