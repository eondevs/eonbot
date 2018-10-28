package config

import (
	"encoding/json"
	"eonbot/pkg/file"
	"eonbot/pkg/strategy"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	stratSuffix = "strat"
)

var (
	ErrInvalidStratSuffix = fmt.Errorf("strategy file name must have '-%s' suffix", stratSuffix)
)

type Strateger interface {
	Get(name string) strategy.Strategy
	Set(name string, strat strategy.Strategy)
	GetAll() map[string]strategy.Strategy
	IsStratChanged(name string) ChangeStatus
	UpdateStratCheckTime(name string)
	removeStrat(name string)

	multipleCommons
}

type strategies struct {
	confUtils
	strats map[string]*stratHolder
}

type stratHolder struct {
	confMod
	lastChangeCheck
	strat strategy.Strategy
}

func (s *stratHolder) isChanged() bool {
	res := s.confMod.modTime().After(s.lastChangeCheck.lastCheck())
	return res
}

func newStrategies(exec ExecConfiger) *strategies {
	return &strategies{
		confUtils: confUtils{exec: exec},
		strats:    make(map[string]*stratHolder),
	}
}

func (s *strategies) Get(name string) (strat strategy.Strategy) {
	str := s.strats[name]
	if str != nil {
		strat = str.strat
	}
	return strat
}

func (s *strategies) Set(name string, strat strategy.Strategy) {
	s.Lock()
	str, ok := s.strats[name]
	if ok { // we need to preserve last check time
		str.setModTime(time.Now().UTC())
		str.strat = strat
	} else {
		newStr := &stratHolder{strat: strat}
		newStr.wasLoaded()
		newStr.setModTime(time.Now().UTC())
		s.strats[name] = newStr
	}
	s.Unlock()
	if !s.IsLoaded() {
		s.wasLoaded()
	}
}

func (s *strategies) GetAll() (strats map[string]strategy.Strategy) {
	strats = make(map[string]strategy.Strategy)
	s.RLock()
	for k, str := range s.strats {
		strats[k] = str.strat
	}
	s.RUnlock()
	return strats
}

func (s *strategies) IsStratChanged(name string) ChangeStatus {
	s.Lock()
	str, ok := s.strats[name]
	if ok {
		if str.isChanged() {
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

func (s *strategies) UpdateStratCheckTime(name string) {
	s.Lock()
	strat, ok := s.strats[name]
	if ok {
		strat.updateLastCheck()
	}
	s.Unlock()
}

func (s *strategies) removeStrat(name string) {
	s.Lock()
	delete(s.strats, name)
	s.Unlock()
}

/*
   commons
*/

func (s *strategies) SetAndSaveJSON(d []byte) error {
	var data struct {
		FileName string          `json:"fileName"`
		Strategy json.RawMessage `json:"strategy"`
	}
	if err := json.Unmarshal(d, &data); err != nil {
		return s.annErr(err)
	}

	if !file.HasSuffix(data.FileName, "-", stratSuffix) {
		return s.annErr(ErrInvalidStratSuffix)
	}

	name := file.RemoveSuffix(data.FileName, "-", stratSuffix)

	var strat strategy.Strategy
	if err := strat.SetName(name); err != nil {
		return err
	}
	if err := json.Unmarshal(data.Strategy, &strat); err != nil {
		return s.annErr(err)
	}

	p := file.JSONPath(s.exec.Get().StrategiesDir, data.FileName)
	if err := file.SaveJSONBytes(p, data.Strategy); err != nil {
		return s.annErr(err)
	}

	s.Set(name, strat)
	s.setModTime(time.Now().UTC())

	return nil
}

func (s *strategies) Load(strict bool) error {
	p := s.exec.Get().StrategiesDir
	if err := file.RequiredExists(p, true); err != nil {
		if strict {
			return s.annErr(err)
		}

		return nil
	}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return s.annErr(err)
	}

	if files == nil || len(files) <= 0 {
		if strict {
			return s.annErr(errors.New("directory is empty"))
		}

		return nil
	}

	var mod bool

	for _, f := range files {
		if !file.IsJSONExt(f.Name()) || !file.HasSuffix(f.Name(), "-", stratSuffix) {
			continue
		}

		if !f.ModTime().After(s.modTime()) {
			continue
		}

		name := file.RemoveSuffix(f.Name(), "-", stratSuffix)

		strat := strategy.Strategy{}
		if err := strat.SetName(name); err != nil {
			return err
		}
		if err := file.LoadJSON(path.Join(p, f.Name()), &strat); err != nil {
			return s.annErr(err)
		}

		s.Set(name, strat)
		if !mod {
			mod = true
		}
	}

Outer:
	for stratFile := range s.GetAll() {
		for _, f := range files {
			if !file.IsJSONExt(f.Name()) || !file.HasSuffix(f.Name(), "-", stratSuffix) {
				continue
			}

			if file.RemoveSuffix(f.Name(), "-", stratSuffix) == stratFile {
				continue Outer
			}
		}

		s.removeStrat(stratFile)
		if !mod {
			mod = true
		}
	}

	if mod {
		s.setModTime(time.Now().UTC())
	}

	return nil
}

func (s *strategies) Download(fileN string) ([]byte, error) {
	if !file.HasSuffix(fileN, "-", stratSuffix) {
		return nil, s.annErr(ErrInvalidStratSuffix)
	}

	res, err := file.Load(file.JSONPath(s.exec.Get().StrategiesDir, fileN))
	if err != nil {
		return nil, s.annErr(err)
	}
	return res, nil
}

func (s *strategies) Remove(fileN string) error {
	if !file.HasSuffix(fileN, "-", stratSuffix) {
		return s.annErr(ErrInvalidStratSuffix)
	}

	if s.onlyOneLeft() {
		return s.annErr(errors.New("only one strategy is left, upload another one to be able to issue removal commands"))
	}

	p := file.JSONPath(s.exec.Get().StrategiesDir, fileN)
	if err := file.Remove(p); err != nil {
		return s.annErr(err)
	}

	s.removeStrat(fileN)
	s.setModTime(time.Now().UTC())
	return nil
}

/*
   internal
*/

func (s *strategies) onlyOneLeft() (res bool) {
	s.Lock()
	res = len(s.strats) == 1
	s.Unlock()
	return res
}

// annErr annotates and wraps all
// errors returned by this type.
func (s *strategies) annErr(err error) error {
	return errors.Wrap(err, "strategies configs manager")
}
