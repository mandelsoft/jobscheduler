package condition

import (
	"sync"

	"github.com/mandelsoft/goutils/optionutils"
)

type ExplicitCondition struct {
	lock    sync.Mutex
	enabled bool
	valid   bool
	final   bool
	trigger func()
}

var _ Condition = Explicit()

func Explicit() *ExplicitCondition {
	return &ExplicitCondition{}
}

func (e *ExplicitCondition) SetStateTrigger(t func()) {
	e.trigger = t
}

func (e *ExplicitCondition) _trigger() {
	if e.trigger != nil {
		e.trigger()
	}
}

func (e *ExplicitCondition) Enable(b ...bool) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.final {
		panic("explicit condition is final")
	}
	e.enabled = optionutils.BoolOption(b...)
	e.final = true
	e.valid = true
	e._trigger()
}

func (e *ExplicitCondition) SetEnabled(enabled bool) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.final {
		panic("explicit condition is final")
	}
	e.enabled = enabled
	e._trigger()
}

func (e *ExplicitCondition) SetValid() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.valid = true
	e._trigger()
}

func (e *ExplicitCondition) SetFinal() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.final = true
	e._trigger()
}

func (e *ExplicitCondition) IsEnabled() bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.enabled && e.valid
}

func (e *ExplicitCondition) GetState() State {
	e.lock.Lock()
	defer e.lock.Unlock()
	return State{e.enabled, e.final, e.valid}
}

func (e *ExplicitCondition) Evaluate(event Event) {
}

func (e *ExplicitCondition) Walk(walker Walker) bool {
	return walker.Walk(e)
}
