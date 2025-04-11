package condition

import (
	"slices"
)

type subCondition struct {
	self  Condition
	sub   []Condition
	def   bool
	force func(sub State, result *State)
}

func (t *subCondition) Evaluate(event Event) {
	for _, n := range t.sub {
		n.Evaluate(event)
	}
}

func (t *subCondition) GetState() State {
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

func (t *subCondition) IsEnabled() bool {
	s := t.GetState()
	return s.Valid && s.Enabled
}

func (t *subCondition) Walk(w Walker) bool {
	if !w.Walk(t.self) {
		return false
	}
	for _, n := range t.sub {
		if !n.Walk(w) {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

type _And struct {
	subCondition
}

func And(t ...Condition) Condition {
	n := &_And{subCondition{sub: slices.Clone(t), def: true}}
	n.self = n
	n.force = n.check
	return n
}

func (t *_And) check(sub State, result *State) {
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

type _Or struct {
	subCondition
}

func Or(t ...Condition) Condition {
	n := &_Or{subCondition{sub: slices.Clone(t), def: false}}
	n.self = n
	n.force = n.check
	return n
}

func (t *_Or) check(sub State, result *State) {
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

type _Not struct {
	subCondition
}

func Not(t Condition) Condition {
	n := &_Not{subCondition{sub: []Condition{t}, def: true}}
	n.self = n
	n.force = n.check
	return n
}

func (t *_Not) check(sub State, result *State) {
	result.Enabled = !sub.Enabled
}
