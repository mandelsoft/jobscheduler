package scheduler

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/ctxutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

type stateJobs interface {
	Add(*job)
	Remove(*job)
	State() State
}

type readyState struct {
	processors.Queue[job, *job]
}

func (s *readyState) State() State {
	return READY
}

type generalState struct {
	lock  sync.Mutex
	set   set.Set[*job]
	state State
}

func newState(state State) *generalState {
	return &generalState{
		set:   set.New[*job](),
		state: state,
	}
}

func (s *generalState) State() State {
	return s.state
}

func (s *generalState) Add(j *job) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.set.Add(j)
}

func (s *generalState) Remove(j *job) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.set.Delete(j)
}

func (s *generalState) Elements() iter.Seq[*job] {
	return func(yield func(*job) bool) {
		for v := range maps.Clone(s.set) {
			if !yield(v) {
				return
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

var schedulerAttr = ctxutils.NewAttribute[Scheduler]()

func GetScheduler(ctx context.Context) Scheduler {
	return schedulerAttr.Get(ctx)
}

func setScheduler(ctx context.Context, scheduler Scheduler) context.Context {
	return schedulerAttr.Set(ctx, scheduler)
}

type scheduler struct {
	lock   synclog.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	name       string
	numRange   int
	jobRange   int
	processors *processors.Processors[*job]
	limiter    processors.Limiter[*job]

	initial   *generalState
	waiting   *generalState
	ready     *readyState
	running   *generalState
	done      *generalState
	discarded *generalState
}

func New(name ...string) Scheduler {
	sn := general.Optional(name...)
	if sn == "" {
		sn = "scheduler" + ctxutils.NewId()
	} else {
		sn = "scheduler " + sn
	}
	q, l := processors.NewQueueWithName[job](sn, func(j *job) string { return j.id })
	s := &scheduler{
		name: sn,
		lock: synclog.NewMutex(sn),

		initial:   newState(INITIAL),
		ready:     &readyState{q},
		waiting:   newState(WAITING),
		running:   newState(RUNNING),
		done:      newState(DONE),
		discarded: newState(DISCARDED),
		limiter:   l,
	}
	s.processors = processors.NewProcessors[*job](s.create, s.limiter)
	return s
}

func (s *scheduler) GetName() string {
	return s.name
}

func (s *scheduler) GetPool() processors.Pool {
	return s.processors
}

func (s *scheduler) Cancel() {
	s.ready.Monitor().Lock()
	defer s.ready.Monitor().Unlock()

	s.processors.Cancel()
}

func (s *scheduler) Wait() {
	s.processors.Wait()
}

func (s *scheduler) AddProcessor() {
	s.processors.New()
}

func (s *scheduler) RemoveProcessor(ctx context.Context) {
	s.processors.Discard(ctx)
}

func (s *scheduler) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.processors.Run(setScheduler(ctx, s))
}

func (s *scheduler) create(id int) processors.Runner {
	return &processor{
		id:        id,
		scheduler: s,
	}
}

func (s *scheduler) Raise(evt condition.Event) {
	for j := range s.waiting.Elements() {
		if j.definition.discard != nil {
			j.definition.trigger.Evaluate(evt)
			js := j.definition.discard.GetState()
			if js.Valid {
				if js.Enabled {
					j.SetState(s.discarded)
				}
			}
		}
	}
	for j := range s.waiting.Elements() {
		if j.definition.trigger != nil {
			j.definition.trigger.Evaluate(evt)
			js := j.definition.trigger.GetState()
			if js.Valid {
				if js.Enabled {
					j.SetState(s.ready)
				} else {
					if js.Final {
						j.SetState(s.discarded)
					}
				}
			}
		}
	}
}

func (s *scheduler) IsStarted() bool {
	return s.processors.IsStarted()
}

func (s *scheduler) Apply(def JobDefinition) (Job, error) {
	if !s.IsStarted() {
		return nil, fmt.Errorf("not started")
	}

	s.lock.Lock()
	s.jobRange++
	id := fmt.Sprintf("%s[%d]", def.name, s.jobRange)
	s.lock.Unlock()

	j := &job{
		lock:       synclog.NewMutex(fmt.Sprintf("job %s", id)),
		id:         id,
		scheduler:  s,
		definition: def,
		state:      nil,
		err:        nil,
		result:     nil,
	}

	for _, h := range def.handlers {
		j.RegisterHandler(h)
	}
	j.SetState(s.initial)
	return j, nil
}
