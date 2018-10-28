package pkg

import (
	"encoding/json"
	"eonbot/pkg/strategy"
	"errors"
	"time"
)

// StreamCycle contains single cycle execution info.
type StreamCycle struct {
	// StartedAt specifies timestamp when the cycle was started.
	StartedAt time.Time `json:"startedAt"`

	// CompletedAt specifies timestamp when the cycle was completed.
	CompletedAt time.Time `json:"completedAt"`

	// IsSuccessful specifies whether the cycle was successful or
	// resulted in error. If successful, Result field will be filled,
	// if unsuccessful, Error field will be filled.
	IsSuccessful bool `json:"isSuccessful"`

	// Result specifies cycle result data.
	// If cycle was unsuccessful, will be empty.
	Result Resulter `json:"result,omitempty"`

	// Error specifies cycle error message.
	// If cycle was successful, will be empty.
	Error string `json:"error,omitempty"`
}

func NewStreamCycle(started, completed time.Time, res Resulter, err error) *StreamCycle {
	isSuccessful := true
	var errMsg string
	if err != nil {
		isSuccessful = false
		errMsg = err.Error()
	}

	return &StreamCycle{
		StartedAt:    started,
		CompletedAt:  completed,
		IsSuccessful: isSuccessful,
		Result:       res,
		Error:        errMsg,
	}
}

func (s *StreamCycle) UnmarshalJSON(d []byte) error {
	var tmp struct {
		StartedAt    time.Time       `json:"startedAt"`
		CompletedAt  time.Time       `json:"completedAt"`
		IsSuccessful bool            `json:"isSuccessful"`
		Result       json.RawMessage `json:"result"`
		Error        string          `json:"error"`
	}

	if err := json.Unmarshal(d, &tmp); err != nil {
		return err
	}

	s.StartedAt = tmp.StartedAt
	s.CompletedAt = tmp.CompletedAt
	s.IsSuccessful = tmp.IsSuccessful
	s.Error = tmp.Error

	// if cycle was unsuccessful i.e. result is not
	// present, return nil and don't unmarshal.
	if !s.IsSuccessful {
		return nil
	}

	var tmpRes struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(tmp.Result, &tmpRes); err != nil {
		return err
	}

	switch tmpRes.Type {
	case StrategiesResType:
		res := &StrategiesResult{}
		if err := json.Unmarshal(tmp.Result, res); err != nil {
			return err
		}
		s.Result = res
	case OpenOrdersResType:
		res := &OpenOrdersResult{}
		if err := json.Unmarshal(tmp.Result, res); err != nil {
			return err
		}
		s.Result = res
	default:
		return errors.New("result type is invalid")
	}

	return nil
}

const (
	StrategiesResType = "strategies"
	OpenOrdersResType = "open-orders"
)

// Resulter is the interface implemented by
// valid stream cycle result types.
type Resulter interface {
	Type() string
}

// ResultCore is a structure that contains
// core fields that all result types should
// have.
type ResultCore struct {
	// ResultType specifies Resulter implementation
	// type.
	ResultType string `json:"type"`
}

// StrategiesResult contains results/snapshots of
// all strategies used by this cycle.
type StrategiesResult struct {
	ResultCore

	// Snapshots specifies a map of each strategy's
	// snapshots.
	Snapshots map[string]strategy.Snapshot `json:"snapshots"`
}

// NewStrategiesResult creates new strategies Resulter
// implementation object.
func NewSrategiesResult(snapshots map[string]strategy.Snapshot) *StrategiesResult {
	if snapshots == nil {
		snapshots = make(map[string]strategy.Snapshot)
	}
	return &StrategiesResult{
		ResultCore{
			ResultType: StrategiesResType,
		},
		snapshots,
	}
}

func (s *StrategiesResult) Type() string {
	return s.ResultType
}

// OpenOrdersResult contains general open orders
// statistic after the cycle.
type OpenOrdersResult struct {
	ResultCore

	// Total specifies how many open orders
	// there were at the beginning of the cycle
	// (before cancelling).
	Total int `json:"total"`

	// Open specifies how many open orders are
	// left not cancelled.
	Open int `json:"open"`

	// Cancelled specifies how many open orders
	// were cancelled during this cycle.
	Cancelled int `json:"cancelled"`
}

// NewOpenOrdersResult creates new open orders Resulter
// implementation object.
func NewOpenOrdersResult(total, open, cancelled int) *OpenOrdersResult {
	return &OpenOrdersResult{
		ResultCore{
			ResultType: OpenOrdersResType,
		},
		total, open, cancelled,
	}
}

func (o *OpenOrdersResult) Type() string {
	return o.ResultType
}
