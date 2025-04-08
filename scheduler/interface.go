package scheduler

import (
	"context"
	"io"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/queue"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type State string

const (
	INITIAL   State = "initial"
	PENDING   State = "pending"
	WAITING   State = "waiting"
	RUNNING   State = "running"
	READY     State = "ready"
	BLOCKED   State = "blocked"
	DONE      State = "done"
	DISCARDED State = "discarded"
)

type Priority = queue.Priority

const DEFAULT_PRIORITY Priority = 100

type Scheduler interface {
	processors.PoolProvider

	AddProcessor(n ...int)
	RemoveProcessor(ctx context.Context)
	SetExtension(e Extension)

	Run(ctx context.Context) error

	JobManager

	Cancel()
	Wait()
}

type JobManager interface {
	Apply(JobDefinition) (Job, error)
}

type Job interface {
	String() string

	GetId() string
	GetScheduler() Scheduler

	GetState() State
	GetResult() (Result, error)
	GetPriority() Priority

	Schedule() error
	Wait()

	GetExtension(typ string) JobExtension

	RegisterHandler(handler EventHandler)
	UnregisterHandler(handler EventHandler)
}

type SchedulingContext interface {
	context.Context
	io.Writer

	Job() Job
}

type Result interface{}

type Runner interface {
	Run(SchedulingContext) (Result, error)
}

type RunnerFunc func(schedulingContext SchedulingContext) (Result, error)

func (f RunnerFunc) Run(s SchedulingContext) (Result, error) {
	return f(s)
}

type EventHandler interface {
	HandleJobEvent(event JobEvent)
}

type JobDefinition struct {
	name      string
	runner    Runner
	trigger   condition.Condition
	discard   condition.Condition
	priority  Priority
	handlers  []EventHandler
	extension ExtensionDefinition
}

func DefineJob(name string, runner ...Runner) JobDefinition {
	return JobDefinition{name: name, runner: general.Optional(runner...), priority: DEFAULT_PRIORITY}
}

func (d JobDefinition) GetName() string {
	return d.name
}

func (d JobDefinition) SetRunner(r Runner) JobDefinition {
	d.runner = r
	return d
}

func (d JobDefinition) SetPriority(p Priority) JobDefinition {
	d.priority = p
	return d
}

func (d JobDefinition) SetCondition(c condition.Condition) JobDefinition {
	d.trigger = c
	return d
}

func (d JobDefinition) SetDiscardCondition(c condition.Condition) JobDefinition {
	d.discard = c
	return d
}

func (d JobDefinition) AddHandler(h EventHandler) JobDefinition {
	d.handlers = sliceutils.CopyAppend(d.handlers, h)
	return d
}

func (d JobDefinition) SetExtension(e ExtensionDefinition) JobDefinition {
	d.extension = e
	return d
}

func (d JobDefinition) GetExtension(typ ...string) ExtensionDefinition {
	if len(typ) == 0 || typ[0] == "" {
		return d.extension
	}
	if d.extension == nil {
		return nil
	}
	return d.extension.GetExtension(typ[0])
}
