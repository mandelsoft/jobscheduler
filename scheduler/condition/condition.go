package condition

import (
	"sync"
)

type ExplicitCondition struct {
	lock    sync.Mutex
	enabled bool
	valid   bool
	final   bool
}

var _ Condition = Explicit()

func Explicit() *ExplicitCondition {
	return &ExplicitCondition{}
}

func (e *ExplicitCondition) SetEnabled(enabled bool) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.final {
		panic("explicit condition is final")
	}
	e.enabled = enabled
}

func (e *ExplicitCondition) SetValid() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.valid = true
}

func (e *ExplicitCondition) SetFinal() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.final = true
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
	return true
}
