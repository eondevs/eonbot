package outcome

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/leebenson/conform"
)

const (
	BuyOutcome      = "buy"
	SellOutcome     = "sell"
	DCAOutcome      = "dca"
	TelegramOutcome = "telegram"
	SandboxOutcome  = "sandbox"
)

type OutcomeType interface {
	Validate() error
	Reset()
}

type Outcome struct {
	Type string
	Raw  []byte
	Conf OutcomeType
}

func (o *Outcome) Clone() (*Outcome, error) {
	out := &Outcome{}
	if err := out.UnmarshalJSON(o.Raw); err != nil {
		return nil, err
	}

	return out, nil
}

func (o *Outcome) UnmarshalJSON(d []byte) error {
	outcome := struct {
		Type       string          `json:"type"`
		Properties json.RawMessage `json:"properties"`
	}{}

	if err := json.Unmarshal(d, &outcome); err != nil {
		return errors.Wrap(err, "outcome")
	}

	outcome.Type = strings.ToLower(outcome.Type)
	var conf OutcomeType
	switch outcome.Type {
	case BuyOutcome:
		var buyConf Buy
		if err := json.Unmarshal(outcome.Properties, &buyConf); err != nil {
			return errors.Wrap(err, "buy outcome")
		}
		conf = &buyConf
	case SellOutcome:
		var sellConf Sell
		if err := json.Unmarshal(outcome.Properties, &sellConf); err != nil {
			return errors.Wrap(err, "sell outcome")
		}
		conf = &sellConf
	case DCAOutcome:
		var dcaConf DCA
		if err := json.Unmarshal(outcome.Properties, &dcaConf); err != nil {
			return errors.Wrap(err, "dca outcome")
		}
		dcaConf.Buy.BasePercent = true
		conf = &dcaConf
	case TelegramOutcome:
		var telegramConf Telegram
		if err := json.Unmarshal(outcome.Properties, &telegramConf); err != nil {
			return errors.Wrap(err, "telegram outcome")
		}
		conf = &telegramConf
	case SandboxOutcome:
		conf = &Sandbox{}
	default:
		return errors.New("outcome: type is invalid")
	}

	conform.Strings(conf)

	if err := conf.Validate(); err != nil {
		return err
	}

	o.Type = outcome.Type
	o.Conf = conf
	o.Raw = d
	return nil
}
