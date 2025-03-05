package scheduler

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/queue"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

type stateJobs interface {
	Add(*job)
	Remove(*job)
	State() State
}

type readyState struct {
	queue.Queue[job, *job]
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

type processorState struct {
	lock      synclog.Mutex
	processor *processor

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func newProcessor(p *processor) *processorState {
	return &processorState{lock: synclog.NewMutex(fmt.Sprintf("processor %d", p.id)), processor: p}
}

func (s *processorState) Run(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ctx, s.cancel = context.WithCancel(ctx)
	s.wg.Add(1)
	go func() {
		s.processor.run(0, ctx)
		s.wg.Done()

		s.processor.scheduler.RemoveProcessor(s.processor)
	}()
}

func (s *processorState) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *processorState) Wait() {
	s.wg.Wait()
}

type scheduler struct {
	lock   synclog.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	name       string
	numRange   int
	jobRange   int
	processors map[*processor]*processorState

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
		sn = "scheduler"
	} else {
		sn = "scheduler " + sn
	}
	return &scheduler{
		name: general.OptionalDefaulted("scheduler", name...),
		lock: synclog.NewMutex(sn),

		processors: map[*processor]*processorState{},

		initial:   newState(INITIAL),
		ready:     &readyState{queue.New[job](func(j *job) string { return j.id })},
		waiting:   newState(WAITING),
		running:   newState(RUNNING),
		done:      newState(DONE),
		discarded: newState(DISCARDED),
	}
}

func (s *scheduler) GetName() string {
	return s.name
}

func (s *scheduler) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		s.cancel()
	}
}

func (s *scheduler) Wait() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, p := range s.processors {
		p.Wait()
	}
}

func (s *scheduler) AddProcessor() *processor {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.numRange++
	p := &processor{
		scheduler: s,
		id:        s.numRange,
	}

	s.processors[p] = newProcessor(p)
	if s.ctx != nil {
		s.processors[p].Run(s.ctx)
	}
	return p
}

func (s *scheduler) RemoveProcessor(p *processor) {
	s.lock.Lock()

	state := s.processors[p]
	if state != nil {
		s.lock.Unlock()

		state.Cancel()
		state.Wait()
	} else {
		s.lock.Unlock()
	}
}

func (s *scheduler) Run(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		return fmt.Errorf("already started")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	for _, p := range s.processors {
		p.Run(s.ctx)
	}
	return nil
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
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.ctx != nil
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
