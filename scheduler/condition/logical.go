package condition

import (
	"slices"
)

type SubCondition struct {
	self  Condition
	sub   []Condition
	def   bool
	force func(sub State, result *State)
}

func (t *SubCondition) Evaluate(event Event) {
	for _, n := range t.sub {
		n.Evaluate(event)
	}
}

func (t *SubCondition) GetState() State {
	if len(t.sub) == 0 {
		return State{true, true, true}
	}

	s := State{t.def, false, false}
	vcnt := 0
	fcnt := 0
	for _, n := range t.sub {
		sub := n.GetState()
		t.force(sub, &s)
		if sub.Valid {
			vcnt++
		}
		if sub.Final {
			fcnt++
		}
	}

	s.Final = s.Final || fcnt == len(t.sub)
	s.Valid = s.Valid || vcnt == len(t.sub)
	return s
}

func (t *SubCondition) IsEnabled() bool {
	s := t.GetState()
	return s.Valid && s.Enabled
}

func (t *SubCondition) Walk(w Walker) bool {
	if t.self.Walk(w) {
		for _, n := range t.sub {
			if !n.Walk(w) {
				return false
			}
		}
		return true
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

type AndCondition struct {
	SubCondition
}

func And(t ...Condition) Condition {
	n := &AndCondition{SubCondition{sub: slices.Clone(t), def: true}}
	n.self = n
	n.force = n.check
	return n
}

func (t *AndCondition) check(sub State, result *State) {
	if !sub.Enabled {
		result.Enabled = false
		if sub.Valid {
			result.Valid = true
		}
		if sub.Final {
			result.Final = true
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type OrCondition struct {
	SubCondition
}

func Or(t ...Condition) Condition {
	n := &OrCondition{SubCondition{sub: slices.Clone(t), def: false}}
	n.self = n
	n.force = n.check
	return n
}

func (t *OrCondition) check(sub State, result *State) {
	if sub.Enabled {
		result.Enabled = true
		if sub.Valid {
			result.Valid = true
		}
		if sub.Final {
			result.Final = true
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type NotCondition struct {
	SubCondition
}

func Not(t Condition) Condition {
	n := &NotCondition{SubCondition{sub: []Condition{t}, def: true}}
	n.self = n
	n.force = n.check
	return n
}

func (t *NotCondition) check(sub State, result *State) {
	result.Enabled = !sub.Enabled
}
