package strategy

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
)

type container struct {
	tool *Tool
	seq  *sequence
}

func newContTool(tool *Tool) *container {
	return &container{
		tool: tool,
	}
}

func newContSeq(seq *sequence) *container {
	return &container{
		seq: seq,
	}
}

func (c *container) isSeq() bool {
	return c.seq != nil
}

func (c *container) isTool() bool {
	return c.tool != nil
}

func (c *container) isUndefined() bool {
	return !c.isTool() && !c.isSeq()
}

func (c *container) isUndefinedErr() error {
	if c.isUndefined() {
		return errors.New("sequence logic block type invalid")
	}

	return nil
}

func (c *container) isBoth() bool {
	return c.isTool() && c.isSeq()
}

func (c *container) isBothErr() error {
	if c.isBoth() {
		return errors.New("sequence logic block type invalid")
	}

	return nil
}

func (c *container) clone() (*container, error) {
	cont := &container{}

	if c.tool != nil {
		tool, err := c.tool.clone()
		if err != nil {
			return nil, err
		}
		cont.tool = tool
	}

	if c.seq != nil {
		seq, err := c.seq.clone()
		if err != nil {
			return nil, err
		}
		cont.seq = seq
	}

	return cont, nil
}

// candlesCount returns the amount of candles
// needed to satisfy all of the inner-containers
// and their tools.
func (c *container) candlesCount() int {
	if c.isTool() {
		return c.tool.Properties.CandlesCount()
	}

	if c.isSeq() {
		return c.seq.candlesCount()
	}

	return 0
}

// validate cheks all tools in all inner-containers
// for configs errors, etc.
func (c *container) validate() error {
	if err := c.isUndefinedErr(); err != nil {
		return err
	}

	if err := c.isBothErr(); err != nil {
		return err
	}

	if c.isSeq() {
		for _, s := range c.seq.elems {
			if err := s.cont.validate(); err != nil {
				return err
			}
		}
		return nil
	}

	// at this point we know that tool is not nil
	return c.tool.Properties.Validate()
}

func (c *container) snapshot() map[string]tools.FullSnapshot {
	res := make(map[string]tools.FullSnapshot)
	if c.isTool() {
		res[c.tool.ID] = c.tool.Properties.Snapshot().Full(c.tool.RawProperties).SetType(c.tool.Type)
	}

	if c.isSeq() {
		res = c.seq.snapshot()
	}

	return res
}

func (c *container) conditionsMet(d exchange.Data) (bool, error) {
	if err := c.isUndefinedErr(); err != nil {
		return false, err
	}

	if err := c.isBothErr(); err != nil {
		return false, err
	}

	if c.isSeq() {
		return c.seq.conditionsMet(d)
	}

	return c.tool.Properties.ConditionsMet(d)
}

func (c *container) reset() {
	if c.isSeq() {
		c.seq.reset()
	}

	if c.isTool() {
		c.tool.Properties.Reset()
	}
}
