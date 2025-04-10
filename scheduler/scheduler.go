package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/ctxutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

////////////////////////////////////////////////////////////////////////////////

var schedulerAttr = ctxutils.NewAttribute[Scheduler]()

func GetScheduler(ctx context.Context) Scheduler {
	return schedulerAttr.Get(ctx)
}

func setScheduler(ctx context.Context, scheduler Scheduler) context.Context {
	return schedulerAttr.Set(ctx, scheduler)
}

type scheduler struct {
	lock      synclog.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	extension Extension

	name       string
	numRange   atomic.Uint64
	jobRange   atomic.Uint64
	processors *processors.Processors[*job]
	limiter    processors.Limiter[*job]

	initial   *generalState
	waiting   *generalState
	pending   *pendingState
	running   *generalState
	ready     *generalState
	blocked   *generalState
	zombie    *generalState
	done      *finalState
	failed    *finalState
	discarded *finalState
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
		name:      sn,
		lock:      synclog.NewMutex(sn),
		extension: newDefaultExtension(),
		initial:   newState(INITIAL),
		pending:   &pendingState{q},
		waiting:   newState(WAITING),
		running:   newState(RUNNING),
		ready:     newState(READY),
		blocked:   newState(BLOCKED),
		zombie:    newState(ZOMBIE),
		done:      newFinalState(DONE),
		failed:    newFinalState(FAILED),
		discarded: newFinalState(DISCARDED),
		limiter:   l,
	}
	s.processors = processors.NewProcessors[*job](s.create, s.limiter)
	s.processors.SetStateHandler(&stateHandler{s})
	return s
}

func (s *scheduler) SetExtension(e Extension) {
	s.extension = e
}

func (s *scheduler) GetName() string {
	return s.name
}

func (s *scheduler) GetPool() processors.Pool {
	return s.processors
}

func (s *scheduler) Cancel() {
	s.pending.Monitor().Lock()
	defer s.pending.Monitor().Unlock()

	s.processors.Cancel()
}

func (s *scheduler) Wait() {
	s.processors.Wait()
}

func (s *scheduler) AddProcessor(n ...int) {
	if len(n) == 0 {
		s.processors.New()
	} else {
		for _, c := range n {
			for i := 0; i < c; i++ {
				s.processors.New()
			}
		}
	}
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
					j.SetState(s.pending)
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

func (s *scheduler) Apply(def JobDefinition, parent ...Job) (Job, error) {
	var err error

	if !s.IsStarted() {
		return nil, fmt.Errorf("not started")
	}

	var p *job
	var pi Job // avoid types nil pointer (I love go)
	if pi = general.Optional(parent...); pi != nil {
		p = pi.(*job)
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	n := s.jobRange.Add(1)
	id := fmt.Sprintf("%s[%d]", def.name, n)

	ext, err := s.extension.JobExtension(id, def, pi)
	if err != nil {
		return nil, err
	}

	j := &job{
		lock:       synclog.NewMutex(fmt.Sprintf("job %s", id)),
		id:         id,
		scheduler:  s,
		definition: def,
		state:      nil,
		err:        nil,
		result:     nil,
		extension:  ext,
		parent:     p,
		writer:     ext.Writer(),
	}
	if p != nil {
		p.children = append(p.children, j)
	}

	for _, h := range def.handlers {
		j.RegisterHandler(h)
	}
	j.SetState(s.initial)
	return j, nil
}

type stateHandler struct {
	scheduler *scheduler
}

var _ processors.StateHandler = (*stateHandler)(nil)

func (s *stateHandler) Ready(ctx context.Context) {
	GetJob(ctx).(*job).SetState(s.scheduler.ready)
}

func (s *stateHandler) Running(ctx context.Context) {
	GetJob(ctx).(*job).SetState(s.scheduler.running)
}

func (s *stateHandler) Block(ctx context.Context) {
	GetJob(ctx).(*job).SetState(s.scheduler.blocked)
}
