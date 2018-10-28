package settings

import (
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/leebenson/conform"
)

type Remote struct {
	// ExchangeDriverAddress specifies address to which all exchange related
	// requests should be sent.
	ExchangeDriverAddress string `json:"exchangeDriverAddress" conform:"trim"`

	// Telegram contains telegram specific settings.
	Telegram Telegram `json:"telegram"`

	// Internal contains internal RC specific settings.
	Internal Internal `json:"internal"`
}

func (r *Remote) UnmarshalJSON(d []byte) error {
	type TmpRemote Remote

	var tmp TmpRemote

	if err := json.Unmarshal(d, &tmp); err != nil {
		return r.annErr(err)
	}

	*r = Remote(tmp)

	conform.Strings(r)
	return r.validate()
}

// annErr annotates and wraps all
// errors returned by this type.
func (r Remote) annErr(err error) error {
	return errors.Wrap(err, "remote config")
}

func (r Remote) validate() error {
	if r.ExchangeDriverAddress == "" {
		return r.annErr(errors.New("exchange driver address cannot be empty"))
	}

	if err := r.Internal.validate(); err != nil {
		return r.annErr(err)
	}

	if err := r.Telegram.validate(); err != nil {
		return r.annErr(err)
	}

	return nil
}

type Telegram struct {
	// Enable specifies whether telegram module should be activated or not.
	Enable bool `json:"enable"`

	// Token specifies telegram authentication token.
	Token string `json:"token" conform:"trim"`

	// Owner specifies telegram username who will be able to interact with
	// the bot.
	Owner string `json:"owner" conform:"trim"`
}

func (t Telegram) validate() error {
	if !t.Enable {
		return nil
	}

	if t.Token == "" {
		return errors.New("telegram token cannot be empty")
	}

	if t.Owner == "" {
		return errors.New("telegram owner field cannot be empty")
	}

	return nil
}

type Internal struct {
	Username string `json:"username" conform:"trim"`
	Password string `json:"password" conform:"trim"`
}

func (i Internal) validate() error {
	if i.Username == "" {
		return errors.New("remote control username cannot be empty")
	}

	if i.Password == "" {
		return errors.New("remote control password cannot be empty")
	}

	return nil
}
