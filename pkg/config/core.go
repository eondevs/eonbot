package config

import (
	"eonbot/pkg/asset"
	"eonbot/pkg/settings"

	"github.com/sirupsen/logrus"
)

type Manager interface {
	Configer
	Loader
}

// Loader handles configs and strategies loading.
// NOTE: strategies must always be loaded first, because
// they are used in main and sub configs.
type Loader interface {
	// InitLoad tries to load all
	// present configs (except auth and remote) in non-strict
	// mode (non-existence won't emit errors).
	InitLoad() error

	// CheckAndReload checks whether any of the
	// configs was modified and loads it, returns true
	// if it was modified.
	CheckAndLoad() (bool, error)

	UpdateCheckTime()
}

type Configer interface {
	ExecConfig() ExecConfiger
	MainConfig() MainConfiger
	RemoteConfig() RemoteConfiger
	SubConfigs() SubsConfiger
	Strategies() Strateger

	// PairConfig returns active PairConfig.
	// If no sub config is present - MainConfig's PairConfig and true bool
	// value will be returned. If sub config is present - sub PairConfig
	// and false boolean value will be returned.
	PairConfig(pair asset.Pair) (settings.Pair, bool)

	Summary() Summary
}

type Summary struct {
	Auth       bool     `json:"auth"`
	Main       bool     `json:"main"`
	Remote     bool     `json:"remote"`
	SubConfigs []string `json:"subConfigs"`
	Strategies []string `json:"strategies"`
}

type manager struct {
	exec ExecConfiger

	main MainConfiger

	remote RemoteConfiger

	subs SubsConfiger

	strategies Strateger

	lastChangeCheck

	// first specifies whether
	// it's the first time
	// when bot is loading configs.
	first bool
}

func New(conf settings.Exec) *manager {
	exec := newExec(conf)
	strat := newStrategies(exec)

	return &manager{
		exec:       exec,
		main:       newMain(exec, strat),
		remote:     newRemote(exec),
		subs:       newSubs(exec, strat),
		strategies: strat,
		first:      true,
	}
}

func (m *manager) ExecConfig() ExecConfiger {
	return m.exec
}

func (m *manager) MainConfig() MainConfiger {
	return m.main
}

func (m *manager) RemoteConfig() RemoteConfiger {
	return m.remote
}

func (m *manager) SubConfigs() SubsConfiger {
	return m.subs
}

func (m *manager) Strategies() Strateger {
	return m.strategies
}

func (m *manager) PairConfig(pair asset.Pair) (settings.Pair, bool) {
	if conf, found := m.SubConfigs().Get(pair); found {
		return conf, false
	}

	return m.MainConfig().Get().PairsConfig, true
}

func (m *manager) Summary() Summary {
	sum := Summary{
		Main:       m.MainConfig().IsLoaded(),
		Remote:     m.RemoteConfig().IsLoaded(),
		SubConfigs: make([]string, 0),
		Strategies: make([]string, 0),
	}
	for fileName := range m.SubConfigs().GetAll() {
		sum.SubConfigs = append(sum.SubConfigs, fileName)
	}

	for fileName := range m.Strategies().GetAll() {
		sum.Strategies = append(sum.Strategies, fileName)
	}

	return sum
}

func (m *manager) InitLoad() error {
	if err := m.Strategies().Load(false); err != nil {
		return err
	}

	if err := m.MainConfig().Load(false); err != nil {
		return err
	}

	if err := m.RemoteConfig().Load(true); err != nil {
		return err
	}

	if err := m.SubConfigs().Load(false); err != nil {
		return err
	}

	m.RemoteConfig().UpdateCheckTime()
	m.UpdateCheckTime()

	return nil
}

// returns true if modified
func (m *manager) CheckAndLoad() (bool, error) {
	var gMod bool

	if err := m.Strategies().Load(true); err != nil {
		return false, err
	}

	var stratsChanged bool
	if !m.Strategies().IsLoaded() && m.first || m.Strategies().modTime().After(m.lastCheck()) {
		if !gMod {
			gMod = true
		}
		stratsChanged = true

		if !m.Strategies().modTime().IsZero() {
			logrus.StandardLogger().Debug("strategies loaded")
		}
	}

	if err := m.MainConfig().Load(true); err != nil {
		return false, err
	}

	var mainChanged bool
	if !m.MainConfig().IsLoaded() && m.first || m.MainConfig().modTime().After(m.lastCheck()) {
		if !gMod {
			gMod = true
		}
		mainChanged = true

		if !m.MainConfig().modTime().IsZero() {
			logrus.StandardLogger().Debug("main config loaded")
		}
	}

	if err := m.RemoteConfig().Load(true); err != nil {
		return false, err
	}

	if !m.RemoteConfig().IsLoaded() && m.first || m.RemoteConfig().modTime().After(m.lastCheck()) {
		if !gMod {
			gMod = true
		}

		if !m.RemoteConfig().modTime().IsZero() {
			logrus.StandardLogger().Debug("remote config loaded")
		}
	}

	if err := m.SubConfigs().Load(true); err != nil {
		return false, err
	}

	var subsChanged bool
	if !m.SubConfigs().IsLoaded() && m.first || m.SubConfigs().modTime().After(m.lastCheck()) {
		if !gMod {
			gMod = true
		}
		subsChanged = true

		if !m.SubConfigs().modTime().IsZero() {
			logrus.StandardLogger().Debug("sub configs loaded")
		}
	}

	if stratsChanged && mainChanged {
		if err := m.MainConfig().recheckStrategies(); err != nil {
			return false, err
		}
	}

	if stratsChanged && subsChanged {
		if err := m.SubConfigs().recheckStrategies(); err != nil {
			return false, err
		}
	}

	m.UpdateCheckTime()
	if m.first {
		m.first = false
	}

	return gMod, nil
}

func (m *manager) UpdateCheckTime() {
	m.updateLastCheck()
}
