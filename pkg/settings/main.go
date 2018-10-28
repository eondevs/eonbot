package settings

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Main contains general bot and pair settings. If specific pair sub
// config is present, bot will use it instead of Main.
type Main struct {
	PairsConfig Pair `json:"pairsConfig,required"`
	BotConfig   Bot  `json:"botConfig,required"`
}

func (m *Main) UnmarshalJSON(d []byte) error {
	type TmpMain Main

	var tmp TmpMain

	if err := json.Unmarshal(d, &tmp); err != nil {
		return m.annErr(err)
	}

	if err := tmp.PairsConfig.validate(); err != nil {
		return m.annErr(err)
	}

	if err := tmp.BotConfig.validate(); err != nil {
		return m.annErr(err)
	}

	*m = Main(tmp)

	return nil
}

// annErr annotates and wraps all
// errors returned by this type.
func (m Main) annErr(err error) error {
	return errors.Wrap(err, "main config")
}
