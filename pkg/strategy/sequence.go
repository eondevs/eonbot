package strategy

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// STOP joint is only used internally
	// i.e. users cannot add it to their config files.
	containerJoint_STOP = iota
	containerJoint_AND
	containerJoint_OR
)

// isJointValid checks if joint is usable.
func isJointValid(i int) bool {
	return i == containerJoint_AND || i == containerJoint_OR
}

var (
	toolIDRegexp, _ = regexp.Compile(`^[a-zA-Z0-9_./#+-]*$`)
)

type sequence struct {
	elems []*seqElem // order is important
}

func newRootSequence(seq string, tools map[string]*Tool) (*sequence, error) {
	return seqFromString(seq, tools, true)
}

func (s *sequence) clone() (*sequence, error) {
	seq := &sequence{}
	for _, elem := range s.elems {
		cont, err := elem.cont.clone()
		if err != nil {
			return nil, err
		}
		seq.elems = append(seq.elems, &seqElem{
			cont:         cont,
			joinNextWith: elem.joinNextWith,
		})
	}

	return seq, nil
}

func (s *sequence) validate() error {
	for _, elem := range s.elems {
		if err := elem.cont.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s *sequence) candlesCount() int {
	var h int
	for _, elem := range s.elems {
		candles := elem.cont.candlesCount()
		if candles > h {
			h = candles
		}
	}

	return h
}

func (s *sequence) snapshot() map[string]tools.FullSnapshot {
	res := make(map[string]tools.FullSnapshot)
	for _, elem := range s.elems {
		snaps := elem.cont.snapshot()
		res = concatSnapshots(res, snaps)
	}
	return res
}

func (s *sequence) reset() {
	for _, elem := range s.elems {
		elem.cont.reset()
	}
}

type seqElem struct {
	cont         *container
	joinNextWith int
}

func (s *seqElem) filled() bool {
	return s.cont != nil && isJointValid(s.joinNextWith)
}

func (s *seqElem) onlyCont() bool {
	return s.cont != nil && !isJointValid(s.joinNextWith)
}

type bracketInfo struct {
	inside  bool
	skip    int
	content strings.Builder
}

func (b *bracketInfo) add(s string) {
	if b.content.Len() > 0 {
		b.content.WriteString(" ")
	}
	b.content.WriteString(s)
}

func (b *bracketInfo) reset() {
	b.inside = false
	b.skip = 0
	b.content.Reset()
}

func seqFromString(seq string, tools map[string]*Tool, first bool) (*sequence, error) {
	if seq == "" {
		if first {
			return nil, errors.New("sequence cannot be empty")
		}
		return nil, errors.New("sequence inner logic block cannot be empty")
	}

	if tools == nil || len(tools) <= 0 {
		return nil, errors.New("tools list cannot be empty")
	}

	ss := strings.Split(seq, " ")

	seqElems := make([]*seqElem, 0)

	var curr seqElem

	checkKey := func(k string) error {
		if seqElems != nil && len(seqElems) > 0 {
			if curr.cont == nil {
				return fmt.Errorf("reserved '%s' keyword/sign cannot go right after another reserved keyword/sign in the sequence", k)
			}
		} else {
			if curr.cont == nil {
				return fmt.Errorf("sequence cannot start with a reserved '%s' keyword/sign", k)
			}
		}
		return nil
	}

	finally := func() {
		nCurr := curr
		seqElems = append(seqElems, &nCurr)
		curr = seqElem{}
	}

	var brc bracketInfo
	for i, s := range ss {
		switch s {
		case "{":
			if curr.onlyCont() {
				return nil, errors.New("inner logic block must be separated with a joint from previous tool ID or logic block")
			}

			if !brc.inside {
				brc.inside = true
				break
			}
			brc.skip++
			brc.add(s)
		case "}":
			if !brc.inside {
				return nil, errors.New("unexpected closing bracket")
			}
			if brc.skip > 0 {
				brc.skip--
				brc.add(s)
				break
			}
			innerSeq, err := seqFromString(brc.content.String(), tools, false)
			if err != nil {
				return nil, err
			}
			curr.cont = newContSeq(innerSeq)
			brc.reset()
		case "and", "AND", "&", "&&":
			if brc.inside {
				brc.add(s)
				break
			}
			if err := checkKey(s); err != nil {
				return nil, err
			}
			curr.joinNextWith = containerJoint_AND
		case "or", "OR", "|", "||":
			if brc.inside {
				brc.add(s)
				break
			}
			if err := checkKey(s); err != nil {
				return nil, err
			}
			curr.joinNextWith = containerJoint_OR
		default:
			if brc.inside {
				brc.add(s)
				break
			}

			if curr.onlyCont() {
				return nil, errors.New("tool IDs must separated by joints")
			}

			if !toolIDRegexp.MatchString(s) {
				return nil, errors.New("sequence contains invalid symbol(s)")
			}

			tl, exists := tools[s]
			if !exists {
				return nil, fmt.Errorf("'%s' tool ID specified in sequence point to a tool that does not exist", s)
			}

			if tl.IsAssigned() {
				return nil, fmt.Errorf("'%s' tool ID cannot be used in the sequence more than once", s)
			}

			tl.MakeAssigned()
			curr.cont = newContTool(tl)
		}

		if i == len(ss)-1 {
			if curr.joinNextWith != containerJoint_STOP {
				if first {
					return nil, errors.New("sequence cannot end with a reserved keyword/sign")
				}
				return nil, errors.New("sequence inner logic block cannot end with a reserved keyword/sign")
			}

			if brc.inside {
				return nil, errors.New("closing bracket is expected but not found")
			}
		}

		if curr.filled() {
			finally()
		}
	}

	brc.reset()
	finally()

	return &sequence{
		elems: seqElems,
	}, nil
}

func (s *sequence) conditionsMet(d exchange.Data) (bool, error) {
	if s.elems == nil || len(s.elems) <= 0 {
		return false, errors.New("conditions list is empty")
	}

	// conditions separated by OR divided into different validation rows
	// each row of 'conds' represents a set of containers/tools that should return true at the same time
	conds := [][]*container{{}}
	for i, elem := range s.elems {
		conds[len(conds)-1] = append(conds[len(conds)-1], elem.cont)
		if elem.joinNextWith == containerJoint_OR {
			conds = append(conds, make([]*container, 0))
		}

		if i == len(s.elems)-1 {
			if elem.joinNextWith != containerJoint_STOP {
				return false, errors.New("conditions list cannot end with a reserved keyword/sign")
			}
		}
	}

	// interates through *ALL* conditions, doesn't matter if OR is used.
	var ok bool
	for _, cond := range conds {
		success := true
		if cond != nil || len(cond) <= 0 {
			for _, cont := range cond {
				ok1, err := cont.conditionsMet(d)
				if err != nil {
					return false, err
				}

				if !ok1 && success {
					success = false
				}
			}
		} else { // impossible case because we already check if seq slice is not nil, but still
			success = false
		}

		if success && !ok {
			ok = true
		}
	}

	return ok, nil
}
