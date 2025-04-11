package scheduler

import (
	"iter"
	"maps"
	"sync"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/processors"
)

////////////////////////////////////////////////////////////////////////////////

type stateJobs interface {
	Add(*job)
	Remove(*job)
	State() State
}

type pendingState struct {
	processors.Queue[job, *job]
}

func (s *pendingState) State() State {
	return PENDING
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

type finalState struct {
	state State
}

func newFinalState(state State) *finalState {
	return &finalState{
		state: state,
	}
}

func (s *finalState) State() State {
	return s.state
}

func (s *finalState) Add(j *job) {
}

func (s *finalState) Remove(j *job) {
}

func (s *finalState) Elements() iter.Seq[*job] {
	return func(yield func(*job) bool) {
	}
}
