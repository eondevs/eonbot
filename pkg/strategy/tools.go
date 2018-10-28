package strategy

import (
	"encoding/json"
	"errors"
	"strings"

	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	indiBuyPrice "eonbot/pkg/strategy/tools/change/buyprice"
	toolRollerCoaster "eonbot/pkg/strategy/tools/change/rollercoaster"
	toolSimpleChange "eonbot/pkg/strategy/tools/change/simple"
	toolMACD "eonbot/pkg/strategy/tools/oscillators/macd"
	toolRSI "eonbot/pkg/strategy/tools/oscillators/rsi"
	toolStoch "eonbot/pkg/strategy/tools/oscillators/stoch"
	toolTrail "eonbot/pkg/strategy/tools/trends/trailing"
	toolBB "eonbot/pkg/strategy/tools/volatility/bb"
	toolMASpread "eonbot/pkg/strategy/tools/volatility/ma_spread"

	"github.com/leebenson/conform"
)

const (
	testTool       = "test"
	buyprice       = "buyprice"
	simpleChange   = "simplechange"
	rollercoaster  = "rollercoaster"
	rsi            = "rsi"
	macd           = "macd"
	stoch          = "stoch"
	bb             = "bb"
	maSpread       = "maspread"
	trailingTrends = "trailingtrends"
)

type Tool struct {
	ID            string
	Type          string
	RawProperties json.RawMessage
	Properties    ToolProperties

	assigned bool
}

func newToolFromJSON(id, t string, nsProperties json.RawMessage) (*Tool, error) {
	t = strings.ToLower(t)
	properties, err := newToolProperties(t, nsProperties)
	if err != nil {
		return nil, err
	}

	return &Tool{
		ID:            id,
		Type:          t,
		RawProperties: nsProperties,
		Properties:    properties,
	}, nil
}

func (t *Tool) MakeAssigned() {
	t.assigned = true
}

func (t *Tool) IsAssigned() bool {
	return t.assigned
}

func (t *Tool) clone() (*Tool, error) {
	prop, err := newToolProperties(t.Type, t.RawProperties)
	if err != nil {
		return nil, err
	}

	return &Tool{
		ID:            t.ID,
		Type:          t.Type,
		RawProperties: t.RawProperties,
		Properties:    prop,
		assigned:      t.assigned,
	}, nil
}

type ToolProperties interface {
	Validate() error
	ConditionsMet(d exchange.Data) (bool, error)
	CandlesCount() int
	Snapshot() tools.Snapshot
	Reset()
}

func newToolProperties(t string, nsProperties json.RawMessage) (ToolProperties, error) {
	convert := func(target interface{}) error {
		if err := json.Unmarshal(nsProperties, target); err != nil {
			return err
		}
		conform.Strings(target)
		return nil
	}

	switch strings.ToLower(t) {
	case testTool:
		return newToolSpecsMock(convert)
	case buyprice:
		return indiBuyPrice.New(convert)
	case simpleChange:
		return toolSimpleChange.New(convert)
	case rollercoaster:
		return toolRollerCoaster.New(convert)
	case macd:
		return toolMACD.New(convert)
	case rsi:
		return toolRSI.New(convert)
	case stoch:
		return toolStoch.New(convert)
	case trailingTrends:
		return toolTrail.New(convert)
	case bb:
		return toolBB.New(convert)
	case maSpread:
		return toolMASpread.New(convert)
	}
	return nil, errors.New("tool type not recognized")
}

func concatSnapshots(maps ...map[string]tools.FullSnapshot) map[string]tools.FullSnapshot {
	res := make(map[string]tools.FullSnapshot)
	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}

	return res
}

/*
   ToolSpecs mock
*/

type toolPropertiesMock struct {
	conf toolPropertiesMockSettings
}

type toolPropertiesMockSettings struct {
	Err      string         `json:"err"`
	Panic    bool           `json:"panic"`
	Count    int            `json:"count"`
	CondsMet bool           `json:"condsMet"`
	IsReset  bool           `json:"isReset"`
	Snap     tools.Snapshot `json:"snapshot"`
}

func newToolSpecsMock(conv func(target interface{}) error) (*toolPropertiesMock, error) {
	var settings toolPropertiesMockSettings
	if err := conv(&settings); err != nil {
		return nil, err
	}

	return &toolPropertiesMock{
		conf: settings,
	}, nil
}

func (t *toolPropertiesMock) Validate() error {
	if t.conf.Err != "" {
		if t.conf.Panic {
			panic(errors.New(t.conf.Err))
		}
		return errors.New(t.conf.Err)
	}
	return nil
}

func (t *toolPropertiesMock) ConditionsMet(d exchange.Data) (bool, error) {
	if t.conf.Err != "" {
		if t.conf.Panic {
			panic(errors.New(t.conf.Err))
		}
		return false, errors.New(t.conf.Err)
	}
	return t.conf.CondsMet, nil
}

func (t *toolPropertiesMock) CandlesCount() int {
	return t.conf.Count
}

func (t *toolPropertiesMock) Snapshot() tools.Snapshot {
	return t.conf.Snap
}

func (t *toolPropertiesMock) Reset() {
	t.conf.IsReset = true
}
