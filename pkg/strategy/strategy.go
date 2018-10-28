package strategy

import (
	"encoding/json"
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/outcome"
	"eonbot/pkg/strategy/tools"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

const (
	BuyModeStrat  = "buymode"
	SellModeStrat = "sellmode"
	AnyModeStrat  = "anymode"
)

type Strategy struct {
	name       string
	origSeq    string // sequence in original form i.e. string
	seq        *sequence
	outcomes   []*outcome.Outcome
	minCandles int
	stratType  string

	snapshot struct {
		mu       sync.RWMutex
		condsMet bool
	}
}

func (s *Strategy) Clone() (*Strategy, error) {
	seq, err := s.seq.clone()
	if err != nil {
		return nil, err
	}

	outcomes := make([]*outcome.Outcome, 0)
	for _, out := range s.outcomes {
		outClone, err := out.Clone()
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, outClone)
	}

	return &Strategy{
		name:       s.name,
		origSeq:    s.origSeq,
		seq:        seq,
		outcomes:   outcomes,
		minCandles: s.minCandles,
		stratType:  s.stratType,
		snapshot: struct {
			mu       sync.RWMutex
			condsMet bool
		}{condsMet: s.snapshot.condsMet},
	}, nil
}

func (s *Strategy) setCondsMet(v bool) {
	s.snapshot.mu.Lock()
	s.snapshot.condsMet = v
	s.snapshot.mu.Unlock()
}

func (s *Strategy) getCondsMet() (v bool) {
	s.snapshot.mu.RLock()
	v = s.snapshot.condsMet
	s.snapshot.mu.RUnlock()
	return v
}

func (s *Strategy) Name() string {
	return s.name
}

func (s *Strategy) SetName(name string) error {
	if name == "" {
		return errors.New("name cannot empty")
	}

	s.name = name

	return nil
}

func (s *Strategy) ReadyToAct(d exchange.Data) (ready bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch rErr := r.(type) {
			case error:
				err = s.annErr(rErr)
			case string:
				err = s.annErr(errors.New(rErr))
			default:
				err = s.annErr(fmt.Errorf("%v", rErr))
			}
		}
	}()

	ready, err = s.seq.conditionsMet(d)
	if err != nil {
		err = s.annErr(err)
	}
	s.setCondsMet(ready)
	return ready, err
}

func (s *Strategy) Reset(outcomes bool) {
	s.seq.reset()
	if outcomes {
		for _, out := range s.outcomes {
			out.Conf.Reset()
		}
	}
}

func (s *Strategy) Outcomes() []*outcome.Outcome {
	return s.outcomes
}

func (s *Strategy) CandlesNeeded() int {
	return s.minCandles
}

func (s *Strategy) Type() string {
	return s.stratType
}

type Snapshot struct {
	CondsMet bool                          `json:"condsMet"`
	Seq      string                        `json:"seq"`
	Tools    map[string]tools.FullSnapshot `json:"tools"`
}

func (s *Strategy) Snapshot() Snapshot {
	return Snapshot{
		CondsMet: s.getCondsMet(),
		Seq:      s.origSeq,
		Tools:    s.seq.snapshot(),
	}
}

func (s *Strategy) UnmarshalJSON(d []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch rErr := r.(type) {
			case error:
				err = s.annErr(rErr)
			case string:
				err = s.annErr(errors.New(rErr))
			default:
				err = s.annErr(fmt.Errorf("%v", rErr))
			}
		}
	}()

	nsStrategy := struct {
		Seq      string             `json:"seq"`
		Outcomes []*outcome.Outcome `json:"outcomes"`
		Tools    map[string]struct {
			Type       string          `json:"type"`
			Properties json.RawMessage `json:"properties"`
		} `json:"tools"`
	}{}

	if err := json.Unmarshal(d, &nsStrategy); err != nil {
		return s.annErr(err)
	}

	if nsStrategy.Tools == nil || len(nsStrategy.Tools) <= 0 {
		return s.annErr(errors.New("tools list cannot be empty"))
	}

	stratTools := make(map[string]*Tool)
	for k, tl := range nsStrategy.Tools {
		if !toolIDRegexp.MatchString(k) {
			return s.annErr(errors.New("tool list contains ID with invalid symbol(s)"))
		}

		nTl, err := newToolFromJSON(k, tl.Type, tl.Properties)
		if err != nil {
			return s.annErr(err)
		}

		stratTools[k] = nTl
	}

	seq, err := newRootSequence(nsStrategy.Seq, stratTools)
	if err != nil {
		return s.annErr(err)
	}

	if err := seq.validate(); err != nil {
		return s.annErr(err)
	}

	s.outcomes = nsStrategy.Outcomes
	s.origSeq = nsStrategy.Seq
	s.seq = seq
	s.minCandles = s.seq.candlesCount()

	if err := s.determineType(); err != nil {
		return s.annErr(err)
	}

	return nil
}

func (s *Strategy) determineType() error {
	if s.outcomes == nil || len(s.outcomes) <= 0 {
		return errors.New("outcomes list cannot be empty")
	}

	var stratType string
	inUse := make([]string, 0)
	for _, out := range s.outcomes {
		switch out.Type {
		case outcome.BuyOutcome:
			if err := checkOutcomes(inUse, outcome.BuyOutcome); err != nil {
				return err
			}

			stratType = BuyModeStrat
			inUse = append(inUse, outcome.BuyOutcome)
		case outcome.SellOutcome:
			if err := checkOutcomes(inUse, outcome.SellOutcome); err != nil {
				return err
			}

			stratType = SellModeStrat
			inUse = append(inUse, outcome.SellOutcome)
		case outcome.DCAOutcome:
			if err := checkOutcomes(inUse, outcome.DCAOutcome); err != nil {
				return err
			}

			stratType = SellModeStrat
			inUse = append(inUse, outcome.DCAOutcome)
		case outcome.TelegramOutcome:
			if err := checkOutcomes(inUse, outcome.TelegramOutcome); err != nil {
				return err
			}

			if stratType == "" {
				stratType = AnyModeStrat
			}
			inUse = append(inUse, outcome.TelegramOutcome)
		case outcome.SandboxOutcome:
			if err := checkOutcomes(inUse, outcome.SandboxOutcome); err != nil {
				return err
			}

			if stratType == "" {
				stratType = AnyModeStrat
			}
			inUse = append(inUse, outcome.SandboxOutcome)
		default:
			return errors.New("outcome type is invalid")
		}
	}

	s.stratType = stratType

	return nil
}

func checkOutcomes(inUse []string, out string) error {
	if out == outcome.SandboxOutcome && len(inUse) > 0 {
		return errors.New("when using sandbox outcome other outcomes cannot be used")
	}

	for _, o := range inUse {
		if o == outcome.SandboxOutcome { // we're trying to add another element and only one is allowed in sandbox mode
			return errors.New("when using sandbox outcome other outcomes cannot be used")
		}

		if o == out {
			return fmt.Errorf("only one %s outcome can be used per strategy", o)
		}

		switch o {
		case outcome.BuyOutcome:
			if out == outcome.SellOutcome {
				return errors.New("buy and sell outcomes cannot be used in a single strategy")
			} else if out == outcome.DCAOutcome {
				return errors.New("buy and dca outcomes cannot be used in a single strategy")
			}
		case outcome.SellOutcome:
			if out == outcome.BuyOutcome {
				return errors.New("buy and sell outcomes cannot be used in a single strategy")
			} else if out == outcome.DCAOutcome {
				return errors.New("sell and dca outcomes cannot be used in a single strategy")
			}
		case outcome.DCAOutcome:
			if out == outcome.BuyOutcome {
				return errors.New("dca and buy outcomes cannot be used in a single strategy")
			} else if out == outcome.SellOutcome {
				return errors.New("dca and sell outcomes cannot be used in a single strategy")
			}
		}
	}

	return nil
}

// annErr annotates and wraps all
// errors returned by this type.
func (s *Strategy) annErr(err error) error {
	return errors.Wrapf(err, "%s strategy", s.name)
}
