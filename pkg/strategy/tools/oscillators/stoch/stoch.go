package stoch

import (
	"eonbot/pkg/exchange"
	indiStoch "eonbot/pkg/strategy/indicators/stoch"
	"eonbot/pkg/strategy/tools"
)

type Stoch struct {
	stoch    indiStoch.Stoch
	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	indiStoch.StochConfig
	tools.Cond
	tools.Level
}

type snapshot struct {
	indiStoch.StochInfo
}

func New(conf func(v interface{}) error) (*Stoch, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	stoch, err := indiStoch.NewFromConfig(s.StochConfig, 0)
	if err != nil {
		return nil, err
	}

	s.Level.ZeroToHundred()

	return &Stoch{
		stoch: stoch,
		conf:  s,
	}, nil
}

func (s *Stoch) Validate() error {
	if err := s.conf.StochConfig.Validate(); err != nil {
		return err
	}

	if err := s.conf.Cond.Validate(); err != nil {
		return err
	}

	if err := s.conf.Level.Validate(); err != nil {
		return err
	}

	return nil
}

func (s *Stoch) ConditionsMet(d exchange.Data) (bool, error) {
	stoch, err := s.stoch.Calc(d.Candles)
	if err != nil {
		s.snapshot.Clear()
		return false, err
	}

	isMet := s.conf.Cond.Match(stoch.D, s.conf.Level.LevelVal)

	// collect snapshot data
	s.snapshot.Set(snapshot{StochInfo: stoch}, isMet)
	return isMet, nil
}

func (s *Stoch) CandlesCount() int {
	return s.stoch.CandlesCount()
}

func (s *Stoch) Snapshot() tools.Snapshot {
	return s.snapshot.Get()
}

func (s *Stoch) Reset() {}
