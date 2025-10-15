package jobnet

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

func Explicit(name string) Condition {
	return &_Explicit{name: name}
}

type _Explicit struct {
	name string
}

func (e *_Explicit) Explicit() string {
	return e.name
}

func (e *_Explicit) Prepare(conds map[string]condition.Condition) error {
	sample := condition.Explicit()
	c := conds[e.name]
	if c == nil {
		conds[e.name] = sample
		return nil
	}
	if reflect.TypeOf(c) == reflect.TypeOf(sample) {
		return nil
	}
	return fmt.Errorf("inconsistent explicit condition '%s' (%s <> %s)", e.name, reflect.TypeOf(c), reflect.TypeOf(sample))
}

func (e *_Explicit) Create(ctx *NetContext) (condition.Condition, error) {
	c := ctx.Conditions[e.name]
	if c == nil {
		return nil, fmt.Errorf("unknown named condition %q", e.name)
	}
	return c, nil
}

func (_ _Explicit) Validate(m map[string]Job) (set.Set[string], error) {
	return nil, nil
}
