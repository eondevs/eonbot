package settings

import (
	"encoding/json"
	"eonbot/pkg/asset"
	"fmt"

	"github.com/pkg/errors"
)

type Sub struct {
	// Active specifies whether this sub config should be used
	// or not.
	Active bool `json:"active"`

	// Pairs specifies a list of pairs that should use this sub config
	// instead of main config.
	Pairs []asset.Pair `json:"pairs"`

	// PairConfig specifies bot settings for these pairs.
	PairsConfig Pair `json:"pairsConfig"`
}

func (s Sub) validate() error {
	if s.Pairs == nil || len(s.Pairs) <= 0 {
		return s.annErr(errors.New("pairs list cannot be empty"))
	}

	if err := s.checkDupPairs(); err != nil {
		return s.annErr(err)
	}

	if err := s.PairsConfig.validate(); err != nil {
		return s.annErr(err)
	}

	return nil
}

func (s Sub) checkDupPairs() error {
	checked := make([]asset.Pair, 0)
	for _, pair := range s.Pairs {
		for _, checkedPair := range checked {
			if checkedPair.Equal(pair) {
				return fmt.Errorf("%s pair is being used more than once in the sub config pairs list", pair.String())
			}
		}
	}

	return nil
}

func (s *Sub) UnmarshalJSON(d []byte) error {
	type TmpSub Sub

	var tmp TmpSub

	if err := json.Unmarshal(d, &tmp); err != nil {
		return s.annErr(err)
	}

	*s = Sub(tmp)

	return s.validate()
}

// annErr annotates and wraps all
// errors returned by this type.
func (s Sub) annErr(err error) error {
	return errors.Wrap(err, "sub config")
}
