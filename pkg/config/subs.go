package config

import (
	"encoding/json"
	"eonbot/pkg/asset"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"eonbot/pkg/settings"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	subSuffix = "sub"
)

var (
	ErrInvalidSubSuffix = fmt.Errorf("sub-config file name must have '-%s' suffix", subSuffix)
)

type SubsConfiger interface {
	// Get returns currently loaded, active specific pair's PairConfig.
	Get(pair asset.Pair) (settings.Pair, bool)

	// Set sets new SubConfig.
	Set(name string, conf settings.Sub)

	// GetAll returns all currently loaded SubConfigs.
	GetAll() map[string]settings.Sub

	// GetName gets subconfig name by pair.
	GetName(pair asset.Pair) string

	// ValidateInterval checks if the interval is valid.
	ValidateInterval(exch exchange.ExchangeConfirmer) error

	// IsChanged checks if sub config exists and is changed.
	IsSubChanged(pair asset.Pair) ChangeStatus

	// UpdateSubCheckTime updates last check time.
	UpdateSubCheckTime(name string)

	// clear empties sub configs map when directory is empty.
	clear()

	// Remove removes specified element.
	removeSub(name string)

	// recheckStrategies checks if all strategies are valid.
	recheckStrategies() error

	multipleCommons
}

type subs struct {
	confUtils
	confs map[string]*subHolder

	strat Strateger
}

type subHolder struct {
	confMod
	lastChangeCheck
	sub settings.Sub
}

func (s *subHolder) isChanged() bool {
	res := s.confMod.modTime().After(s.lastChangeCheck.lastCheck())
	return res
}

func newSubs(exec ExecConfiger, strat Strateger) *subs {
	return &subs{
		confUtils: confUtils{exec: exec},
		confs:     make(map[string]*subHolder),
		strat:     strat,
	}
}

func (s *subs) Get(pair asset.Pair) (conf settings.Pair, found bool) {
	_, holder := s.get(pair)
	if holder != nil {
		return holder.sub.PairsConfig, true
	} else {
		return settings.Pair{}, false
	}
}

func (s *subs) get(pair asset.Pair) (name string, conf *subHolder) {
	s.RLock()
	for k, holder := range s.confs {
		c := holder.sub
		if !c.Active {
			continue
		}

		var found bool
		for _, curr := range c.Pairs {
			if !curr.Equal(pair) {
				continue
			}

			found = true
			break
		}

		if found {
			conf = holder
			name = k
			break
		}
	}

	s.RUnlock()
	return name, conf
}

func (s *subs) getByName(name string) (conf settings.Sub, exists bool) {
	s.RLock()
	holder, exists := s.confs[name]
	if exists {
		conf = holder.sub
	}
	s.RUnlock()
	return conf, exists
}

func (s *subs) Set(name string, conf settings.Sub) {
	s.Lock()
	sub, ok := s.confs[name]
	if ok { // we need to preserve last check time
		sub.setModTime(time.Now().UTC())
		sub.sub = conf
	} else {
		newSub := &subHolder{sub: conf}
		newSub.wasLoaded()
		newSub.setModTime(time.Now().UTC())
		s.confs[name] = newSub
	}
	s.Unlock()
	if !s.IsLoaded() {
		s.wasLoaded()
	}
}

func (s *subs) GetAll() (confs map[string]settings.Sub) {
	confs = make(map[string]settings.Sub)
	s.RLock()
	for k, sub := range s.confs {
		confs[k] = sub.sub
	}
	s.RUnlock()
	return confs
}

func (s *subs) GetName(pair asset.Pair) string {
	name, _ := s.get(pair)
	return name
}

func (s *subs) ValidateInterval(exch exchange.ExchangeConfirmer) error {
	for _, sub := range s.GetAll() {
		if err := exch.ConfirmInterval(sub.PairsConfig.CandleInterval); err != nil {
			return s.annErr(err)
		}
	}
	return nil
}

func (s *subs) IsSubChanged(pair asset.Pair) ChangeStatus {
	_, holder := s.get(pair)
	s.Lock()
	if holder != nil {
		if holder.isChanged() {
			s.Unlock()
			return Changed
		} else {
			s.Unlock()
			return NotChanged
		}
	}
	s.Unlock()
	return NotExists
}

func (s *subs) UpdateSubCheckTime(name string) {
	s.Lock()
	sub, ok := s.confs[name]
	if ok {
		sub.updateLastCheck()
	}
	s.Unlock()
}

func (s *subs) clear() {
	var cleared bool
	s.Lock()
	if s.confs != nil && len(s.confs) > 0 {
		s.confs = make(map[string]*subHolder, 0)
		cleared = true
	}
	s.Unlock()
	if cleared {
		s.setModTime(time.Now().UTC())
		s.notLoaded()
	}
}

func (s *subs) removeSub(name string) {
	s.Lock()
	delete(s.confs, name)
	s.Unlock()
}

func (s *subs) recheckStrategies() error {
	for _, sub := range s.GetAll() {
		if err := sub.PairsConfig.ValidateStrategies(s.strat.GetAll()); err != nil {
			return err
		}
	}

	return nil
}

/*
   commons
*/

func (s *subs) SetAndSaveJSON(d []byte) error {
	var data struct {
		FileName string          `json:"fileName"`
		Config   json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(d, &data); err != nil {
		return s.annErr(err)
	}

	if !file.HasSuffix(data.FileName, "-", subSuffix) {
		return s.annErr(ErrInvalidSubSuffix)
	}

	var conf settings.Sub
	if err := json.Unmarshal(data.Config, &conf); err != nil {
		return s.annErr(err)
	}

	name := file.RemoveSuffix(data.FileName, "-", subSuffix)
	if err := s.checkIfAdd(conf, name); err != nil {
		return s.annErr(err)
	}

	if err := s.checkStrategies(conf); err != nil {
		return s.annErr(err)
	}

	p := file.JSONPath(s.exec.Get().SubsDir, data.FileName)
	if err := file.SaveJSONBytes(p, data.Config); err != nil {
		return s.annErr(err)
	}

	s.Set(name, conf)
	s.setModTime(time.Now().UTC())

	return nil
}

func (s *subs) Load(strict bool) error {
	p := s.exec.Get().SubsDir
	if !file.Exists(p) {
		s.clear()
		return nil
	}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return s.annErr(err)
	}

	if files == nil || len(files) <= 0 {
		s.clear()
		return nil
	}

	// if sub configs were cleared (i.e. used at some point
	// but removed later), but their timestamps
	// data is non-zero, reset it.
	if !s.IsLoaded() {
		if !s.modTime().IsZero() {
			s.setModTime(time.Unix(0, 0))
		}

		if !s.lastCheck().IsZero() {
			s.resetLastCheck()
		}
	}

	var mod bool

	for _, f := range files {
		if !file.IsJSONExt(f.Name()) || !file.HasSuffix(f.Name(), "-", subSuffix) {
			continue
		}

		if !f.ModTime().After(s.modTime()) {
			continue
		}

		conf := settings.Sub{}
		if err := file.LoadJSON(path.Join(p, f.Name()), &conf); err != nil {
			return s.annErr(err)
		}

		name := file.RemoveSuffix(f.Name(), "-", subSuffix)
		if err := s.checkIfAdd(conf, name); err != nil {
			return s.annErr(err)
		}

		if err := s.checkStrategies(conf); err != nil {
			return s.annErr(err)
		}

		s.Set(name, conf)
		if !mod {
			mod = true
		}
	}

Outer:
	for subFile := range s.GetAll() {
		for _, f := range files {
			if !file.IsJSONExt(f.Name()) || !file.HasSuffix(f.Name(), "-", subSuffix) {
				continue
			}

			if file.RemoveSuffix(f.Name(), "-", subSuffix) == subFile {
				continue Outer
			}
		}
		s.removeSub(subFile)
		if !mod {
			mod = true
		}
	}

	if mod {
		s.setModTime(time.Now().UTC())
	}

	return nil
}

func (s *subs) Download(fileN string) ([]byte, error) {
	if !file.HasSuffix(fileN, "-", subSuffix) {
		return nil, s.annErr(ErrInvalidSubSuffix)
	}

	res, err := file.Load(file.JSONPath(s.exec.Get().SubsDir, fileN))

	if err != nil {
		return nil, s.annErr(err)
	}

	return res, nil
}

func (s *subs) Remove(fileN string) error {
	if !file.HasSuffix(fileN, "-", subSuffix) {
		return s.annErr(ErrInvalidSubSuffix)
	}

	p := file.JSONPath(s.exec.Get().SubsDir, fileN)
	if err := file.Remove(p); err != nil {
		return s.annErr(err)
	}

	s.removeSub(fileN)
	s.setModTime(time.Now().UTC())

	return nil
}

/*
   internal
*/

func (s *subs) checkIfAdd(sub settings.Sub, name string) error {
	// check if we're updating the config or not.
	_, exists := s.getByName(name)
	if exists {
		return nil
	}

	// gather info
	active := make([]asset.Pair, 0)
	for _, c := range s.GetAll() {
		if !c.Active {
			continue
		}

		for _, a := range c.Pairs {
			if len(active) > 0 {
				for _, ac := range active {
					if ac.Equal(a) {
						return fmt.Errorf("%s sub config pair is used more than once (either in one or multiple active subconfigs)", a.String())
					}
				}
			}

			active = append(active, a)
		}
	}

	if !sub.Active {
		return nil
	}

	// check new
	for _, a := range sub.Pairs {
		if len(active) > 0 {
			for _, ac := range active {
				if ac.Equal(a) {
					return fmt.Errorf("%s sub config pair is used more than once (either in one or multiple active subconfigs)", a.String())
				}
			}
		}
		active = append(active, a)
	}

	return nil
}

func (s *subs) checkStrategies(conf settings.Sub) error {
	return conf.PairsConfig.ValidateStrategies(s.strat.GetAll())
}

// annErr annotates and wraps all
// errors returned by this type.
func (s *subs) annErr(err error) error {
	return errors.Wrap(err, "sub configs manager")
}
