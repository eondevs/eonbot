package config

import (
	"encoding/json"
	"eonbot/pkg/file"
	"eonbot/pkg/settings"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	remoteFileName = "remote"
)

type RemoteConfiger interface {
	// Get returns currently loaded RemoteConfig.
	Get() settings.Remote

	// Set sets new RemoteConfig.
	Set(conf settings.Remote)

	singularCommons
}

type remote struct {
	confUtils
	conf settings.Remote
}

func newRemote(exec ExecConfiger) *remote {
	return &remote{confUtils: confUtils{exec: exec}}
}

func (r *remote) Get() (conf settings.Remote) {
	r.RLock()
	conf = r.conf
	r.RUnlock()
	return conf
}

func (r *remote) Set(conf settings.Remote) {
	r.Lock()
	r.conf = conf
	r.Unlock()
	r.setModTime(time.Now().UTC())
	if !r.IsLoaded() {
		r.wasLoaded()
	}
}

/*
   commons
*/

func (r *remote) SetAndSaveJSON(d []byte) error {
	var conf settings.Remote
	if err := json.Unmarshal(d, &conf); err != nil {
		return r.annErr(err)
	}

	p := file.JSONPath(r.exec.Get().ConfigsDir, remoteFileName)
	if err := file.SaveJSONBytes(p, d); err != nil {
		return r.annErr(err)
	}

	r.Set(conf)

	return nil
}

func (r *remote) Load(strict bool) error {
	p := file.JSONPath(r.exec.Get().ConfigsDir, remoteFileName)
	if err := file.RequiredExists(p, false); err != nil {
		if strict {
			return r.annErr(err)
		}

		return nil
	}

	info, err := os.Stat(p)
	if err != nil {
		return r.annErr(err)
	}

	if !info.ModTime().After(r.modTime()) {
		return nil
	}

	conf := settings.Remote{}
	if err := file.LoadJSON(p, &conf); err != nil {
		return r.annErr(err)
	}

	r.Set(conf)
	return nil
}

func (r *remote) Download() ([]byte, error) {
	res, err := file.Load(file.JSONPath(r.exec.Get().ConfigsDir, remoteFileName))
	if err != nil {
		return nil, r.annErr(err)
	}

	return res, nil
}

/*
	internal
*/

// annErr annotates and wraps all
// errors returned by this type.
func (r *remote) annErr(err error) error {
	return errors.Wrap(err, "remote config manager")
}
