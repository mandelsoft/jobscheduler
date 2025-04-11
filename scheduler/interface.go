package scheduler

import (
	"context"
	"io"
	"slices"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/queue"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type State string

const (
	// new unscheduled job
	INITIAL State = "initial"
	// waiting for start condition
	WAITING State = "waiting"
	// waiting to get started
	PENDING State = "pending"
	// processor assigned
	RUNNING State = "running"
	// waiting for processor to continue
	READY State = "ready"
	// started but waitinmg for synchronization operation
	BLOCKED State = "blocked"
	// waiting for uncompleted sub jobs
	ZOMBIE State = "zombie"
	// job completed
	DONE State = "done"
	// job execution failed
	FAILED State = "failed"
	// job not started because of failed start condition
	DISCARDED State = "discarded"
)

func IsFinished(state State) bool {
	switch state {
	case DONE, DISCARDED, FAILED:
		return true
	default:
		return false
	}
}

func IsDone(state State) bool {
	switch state {
	case DONE:
		return true
	default:
		return false
	}
}

func IsDiscarded(state State) bool {
	switch state {
	case DISCARDED:
		return true
	default:
		return false
	}
}

func IsFailed(state State) bool {
	switch state {
	case FAILED:
		return true
	default:
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////

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
	Apply(def JobDefinition, parent ...Job) (Job, error)
	ScheduleDefinition(def JobDefinition, parent ...Job) (Job, error)
	ScheduleDefinitions(defs ...JobDefinition) ([]Job, error)
}

type Job interface {
	String() string

	GetId() string
	GetScheduler() Scheduler

	GetState() State
	GetResult() (Result, error)
	GetPriority() Priority

	Schedule() error
	Cancel()
	Wait()

	GetExtension(typ string) JobExtension

	RegisterHandler(handler EventHandler)
	UnregisterHandler(handler EventHandler)
}

type SchedulingContext interface {
	context.Context
	io.Writer

	Job() Job
	Scheduler() Scheduler
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

type JobDefinition interface {
	GetName() string
	GetRunner() Runner
	GetCondition() condition.Condition
	GetDiscardCondition() condition.Condition
	GetPriority() Priority
	GetHandlers() []EventHandler
	GetExtension(typ ...string) ExtensionDefinition
}

type DefaultJobDefinition struct {
	name      string
	runner    Runner
	trigger   condition.Condition
	discard   condition.Condition
	priority  Priority
	handlers  []EventHandler
	extension ExtensionDefinition
}

var _ JobDefinition = DefaultJobDefinition{}

func DefineJob(name string, runner ...Runner) DefaultJobDefinition {
	return DefaultJobDefinition{name: name, runner: general.Optional(runner...), priority: DEFAULT_PRIORITY}
}

func (d DefaultJobDefinition) GetName() string {
	return d.name
}

func (d DefaultJobDefinition) SetName(name string) DefaultJobDefinition {
	d.name = name
	return d
}

func (d DefaultJobDefinition) GetRunner() Runner {
	return d.runner
}

func (d DefaultJobDefinition) SetRunner(r Runner) DefaultJobDefinition {
	d.runner = r
	return d
}

func (d DefaultJobDefinition) GetPriority() Priority {
	return d.priority
}

func (d DefaultJobDefinition) SetPriority(p Priority) DefaultJobDefinition {
	d.priority = p
	return d
}

func (d DefaultJobDefinition) GetCondition() condition.Condition {
	return d.trigger
}

func (d DefaultJobDefinition) SetCondition(c condition.Condition) DefaultJobDefinition {
	d.trigger = c
	return d
}

func (d DefaultJobDefinition) GetDiscardCondition() condition.Condition {
	return d.discard
}

func (d DefaultJobDefinition) SetDiscardCondition(c condition.Condition) DefaultJobDefinition {
	d.discard = c
	return d
}

func (d DefaultJobDefinition) GetHandlers() []EventHandler {
	return slices.Clone(d.handlers)
}

func (d DefaultJobDefinition) AddHandler(h EventHandler) DefaultJobDefinition {
	d.handlers = sliceutils.CopyAppend(d.handlers, h)
	return d
}

func (d DefaultJobDefinition) SetExtension(e ExtensionDefinition) DefaultJobDefinition {
	d.extension = e
	return d
}

func (d DefaultJobDefinition) GetExtension(typ ...string) ExtensionDefinition {
	if len(typ) == 0 || typ[0] == "" {
		return d.extension
	}
	if d.extension == nil {
		return nil
	}
	return d.extension.GetExtension(typ[0])
}

func newDefinition(def JobDefinition) DefaultJobDefinition {
	return DefaultJobDefinition{
		name:      def.GetName(),
		runner:    def.GetRunner(),
		trigger:   def.GetCondition(),
		discard:   def.GetDiscardCondition(),
		priority:  def.GetPriority(),
		handlers:  def.GetHandlers(),
		extension: def.GetExtension(),
	}
}
