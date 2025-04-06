package scheduler

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/queue"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type State string

const (
	INITIAL   State = "initial"
	READY     State = "ready"
	WAITING   State = "waiting"
	RUNNING   State = "running"
	DONE      State = "done"
	DISCARDED State = "discarded"
)

type Priority = queue.Priority

const DEFAULT_PRIORITY Priority = 100

type Scheduler interface {
	processors.PoolProvider

	AddProcessor()
	RemoveProcessor(ctx context.Context)
	Run(ctx context.Context) error

	Apply(JobDefinition) (Job, error)

	Cancel()
	Wait()
}

type Job interface {
	GetId() string
	GetScheduler() Scheduler

	GetState() State
	GetResult() (Result, error)
	GetPriority() Priority

	Schedule() error
	Wait()

	RegisterHandler(handler EventHandler)
	UnregisterHandler(handler EventHandler)
}

type SchedulingContext struct {
	Scheduler Scheduler
	Pool      processors.Pool
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
	name     string
	runner   Runner
	trigger  condition.Condition
	discard  condition.Condition
	priority Priority
	handlers []EventHandler
}

func DefineJob(name string, runner Runner) JobDefinition {
	return JobDefinition{name: name, runner: runner, priority: DEFAULT_PRIORITY}
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
