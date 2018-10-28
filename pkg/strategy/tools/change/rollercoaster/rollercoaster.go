package rollercoaster

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"errors"
	"github.com/shopspring/decimal"
)

const (
	highestPoint = "highest"
	lowestPoint  = "lowest"
)

type RollerCoaster struct {
	pointVal decimal.Decimal
	conf     settings
	snapshot tools.SnapshotManager
}

type settings struct {
	PointType string `json:"pointType" conform:"trim,lower"`
	tools.CondObject
	tools.Shift
}

type snapshot struct {
	PointVal        decimal.Decimal `json:"pointVal"`
	ShiftedPointVal decimal.Decimal `json:"shiftedPointVal"`
	tools.CondObjectSnapshot
}

func New(conf func(v interface{}) error) (*RollerCoaster, error) {
	var s settings
	if err := conf(&s); err != nil {
		return nil, err
	}

	s.CondObject.AllowTickerPrice()
	s.CondObject.AllowTickerMiscProp()
	s.CondObject.AllowCandlePrice()
	s.CondObject.AllowMA()
	if err := s.CondObject.Init(0); err != nil {
		return nil, err
	}
	return &RollerCoaster{
		conf: s,
	}, nil
}

func (r *RollerCoaster) Validate() error {
	switch r.conf.PointType {
	case highestPoint, lowestPoint:
		break
	default:
		return errors.New("point type is invalid")
	}

	if err := r.conf.CondObject.Validate(); err != nil {
		return err
	}

	if err := r.conf.Shift.Validate(); err != nil {
		return err
	}

	if r.conf.Shift.ShiftVal.Sign() > 0 && r.conf.PointType == highestPoint {
		return errors.New("shift value must be negative when used with highest point type")
	}

	if r.conf.Shift.ShiftVal.Sign() < 0 && r.conf.PointType == lowestPoint {
		return errors.New("shift value must be positive when used with lowest point type")
	}

	return nil
}

func (r *RollerCoaster) ConditionsMet(d exchange.Data) (bool, error) {
	val, err := r.conf.CondObject.Value(d)
	if err != nil {
		r.snapshot.Clear()
		return false, err
	}

	var isMet bool
	var shiftedVal decimal.Decimal
	if r.pointVal.Equal(decimal.Zero) {
		r.pointVal = val
		shiftedVal = r.conf.Shift.CalcVal(r.pointVal)
	} else {
		shiftedVal = r.conf.Shift.CalcVal(r.pointVal)
		switch r.conf.PointType {
		case highestPoint:
			if val.GreaterThan(r.pointVal) {
				r.pointVal = val
				shiftedVal = r.conf.Shift.CalcVal(r.pointVal)
				break
			} else if val.LessThanOrEqual(shiftedVal) {
				isMet = true
			}
		case lowestPoint:
			if val.LessThan(r.pointVal) {
				r.pointVal = val
				shiftedVal = r.conf.Shift.CalcVal(r.pointVal)
				break
			} else if val.GreaterThanOrEqual(shiftedVal) {
				isMet = true
			}
		default:
			r.snapshot.Clear()
			return false, errors.New("point type is invalid")
		}
	}

	// collect snapshot data
	r.snapshot.Set(snapshot{
		PointVal:           r.pointVal,
		ShiftedPointVal:    shiftedVal,
		CondObjectSnapshot: r.conf.Snapshot(val),
	}, isMet)
	return isMet, nil
}

func (r *RollerCoaster) CandlesCount() int {
	return r.conf.CondObject.CandlesCount()
}

func (r *RollerCoaster) Snapshot() tools.Snapshot {
	return r.snapshot.Get()
}

func (r *RollerCoaster) Reset() {
	r.pointVal = decimal.Zero
}
