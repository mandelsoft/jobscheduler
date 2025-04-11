package jobnet

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/scheduler"
)

type EventHandler = scheduler.EventHandler
type ExtensionDefinition = scheduler.ExtensionDefinition
type Priority = scheduler.Priority
type State = scheduler.State

const DEFAULT_PRIORITY = scheduler.DEFAULT_PRIORITY

type Job struct {
	name      string
	runner    Runner
	trigger   Condition
	discard   Condition
	priority  Priority
	handlers  []scheduler.EventHandler
	extension scheduler.ExtensionDefinition
}

type Runner interface {
	CreateRunner(ctx *NetContext) scheduler.Runner
}

type RunnerFunc func(ctx *NetContext) scheduler.Runner

func (r RunnerFunc) CreateRunner(ctx *NetContext) scheduler.Runner {
	return r(ctx)
}

func DefineJob(name string, runner ...Runner) Job {
	return Job{name: name, runner: general.Optional(runner...), priority: DEFAULT_PRIORITY}
}

func (d Job) GetName() string {
	return d.name
}

func (d Job) SetName(name string) Job {
	d.name = name
	return d
}

func (d Job) SetRunner(r Runner) Job {
	d.runner = r
	return d
}

func (d Job) SetPriority(p Priority) Job {
	d.priority = p
	return d
}

func (d Job) SetCondition(c Condition) Job {
	d.trigger = c
	return d
}

func (d Job) SetDiscardCondition(c Condition) Job {
	d.discard = c
	return d
}

func (d Job) AddHandler(h EventHandler) Job {
	d.handlers = sliceutils.CopyAppend(d.handlers, h)
	return d
}

func (d Job) SetExtension(e ExtensionDefinition) Job {
	d.extension = e
	return d
}

func (d Job) GetExtension(typ ...string) ExtensionDefinition {
	if len(typ) == 0 || typ[0] == "" {
		return d.extension
	}
	if d.extension == nil {
		return nil
	}
	return d.extension.GetExtension(typ[0])
}

func (d Job) validate(jobs map[string]Job) (set.Set[string], error) {
	result := errors.ErrListf("job %q", d.name)
	required := set.Set[string]{}

	if d.trigger != nil {
		req, err := d.trigger.Validate(jobs)
		result.Add(err)
		required.AddAll(req)
	}
	return required, result.Result()
}
