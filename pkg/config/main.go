package config

import (
	"encoding/json"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"eonbot/pkg/settings"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	mainFileName = "main"
)

type MainConfiger interface {
	// Get returns currently loaded MainConfig.
	Get() settings.Main

	// Set sets new MainConfig.
	Set(conf settings.Main)

	// Updates active pairs list with new, formatted ones.
	// Won't update mod time.
	UpdateActivePairs(exch exchange.ExchangeConfirmer) error

	// ValidateInterval checks if the interval is valid.
	ValidateInterval(exch exchange.ExchangeConfirmer) error

	// recheckStrategies checks if all strategies are valid.
	recheckStrategies() error

	singularCommons
}

type main struct {
	confUtils
	conf settings.Main

	strat Strateger
}

func newMain(exec ExecConfiger, strat Strateger) *main {
	return &main{
		confUtils: confUtils{exec: exec},
		strat:     strat,
	}
}

func (m *main) Get() (conf settings.Main) {
	m.RLock()
	conf = m.conf
	m.RUnlock()
	return m.conf
}

func (m *main) Set(conf settings.Main) {
	m.Lock()
	m.conf = conf
	m.Unlock()
	m.setModTime(time.Now().UTC())
	if !m.IsLoaded() {
		m.wasLoaded()
	}
}

func (m *main) UpdateActivePairs(exch exchange.ExchangeConfirmer) error {
	pairs, err := exch.ConfirmPairs(m.Get().BotConfig.ActivePairs)
	if err != nil {
		return m.annErr(err)
	}

	m.Lock()
	m.conf.BotConfig.ActivePairs = pairs
	m.Unlock()
	return nil
}

func (m *main) ValidateInterval(exch exchange.ExchangeConfirmer) error {
	err := exch.ConfirmInterval(m.Get().PairsConfig.CandleInterval)
	if err != nil {
		return m.annErr(err)
	}

	return nil
}

func (m *main) recheckStrategies() error {
	err := m.Get().PairsConfig.ValidateStrategies(m.strat.GetAll())
	if err != nil {
		return m.annErr(err)
	}

	return nil
}

/*
   commons
*/

func (m *main) SetAndSaveJSON(d []byte) error {
	var conf settings.Main
	if err := json.Unmarshal(d, &conf); err != nil {
		return m.annErr(err)
	}

	if err := m.checkStrategies(conf); err != nil {
		return m.annErr(err)
	}

	p := file.JSONPath(m.exec.Get().ConfigsDir, mainFileName)
	if err := file.SaveJSONBytes(p, d); err != nil {
		return m.annErr(err)
	}

	m.Set(conf)

	return nil
}

func (m *main) Load(strict bool) error {
	p := file.JSONPath(m.exec.Get().ConfigsDir, mainFileName)
	if err := file.RequiredExists(p, false); err != nil {
		if strict {
			return m.annErr(err)
		}
		return nil
	}

	info, err := os.Stat(p)
	if err != nil {
		return m.annErr(err)
	}

	if !info.ModTime().After(m.modTime()) {
		return nil
	}

	conf := settings.Main{}
	if err := file.LoadJSON(p, &conf); err != nil {
		return m.annErr(err)
	}

	if err := m.checkStrategies(conf); err != nil {
		return m.annErr(err)
	}

	m.Set(conf)
	return nil
}

func (m *main) Download() ([]byte, error) {
	res, err := file.Load(file.JSONPath(m.exec.Get().ConfigsDir, mainFileName))
	if err != nil {
		return nil, m.annErr(err)
	}

	return res, nil
}

/*
   internal
*/

// to use before setting new
func (m *main) checkStrategies(conf settings.Main) error {
	return conf.PairsConfig.ValidateStrategies(m.strat.GetAll())
}

// annErr annotates and wraps all
// errors returned by this type.
func (m *main) annErr(err error) error {
	return errors.Wrap(err, "main config manager")
}
