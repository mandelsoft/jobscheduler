package jobnet

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type Condition interface {
	Prepare(map[string]condition.Condition) error

	Create(ctx *NetContext) (condition.Condition, error)

	Validate(map[string]Job) (set.Set[string], error)
}
