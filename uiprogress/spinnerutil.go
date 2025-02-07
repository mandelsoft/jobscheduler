package uiprogress

import (
	"sync"

	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type RawSpinnerInterface[T any] interface {
	// SetSpeed sets the spinner speed (larger = slower).
	SetSpeed(v int) T

	// SetDone sets the done visualization string.
	SetDone(string) T

	// SetPredefined set predefined spinner phases
	SetPredefined(int) T

	// SetPhases sets the spinner phases
	SetPhases(...string) T
}

type RawSpinner[P ProgressInterface[P]] struct {
	lock sync.Mutex
	self ppi.Self[P, ppi.ProgressProtected[P]]

	phases []string
	speed  int
	done   string

	cnt   int
	phase int
}

var _ RawSpinnerInterface[Spinner] = (*RawSpinner[Spinner])(nil)

func NewRawSpinner[T ProgressInterface[T]](self ppi.Self[T, ppi.ProgressProtected[T]], set int) RawSpinner[T] {
	if set < 0 || SpinnerTypes[set] == nil {
		set = 9
	}
	return RawSpinner[T]{
		self:   self,
		phases: SpinnerTypes[set],
		cnt:    specs.Speed - 1,
		speed:  specs.Speed,
		done:   specs.Done,
	}
}

func (s *RawSpinner[T]) SetSpeed(v int) T {
	s.speed = v
	s.cnt = v - 1
	return s.self.Self()
}

func (s *RawSpinner[T]) SetDone(m string) T {
	s.done = m
	return s.self.Self()
}

func (s *RawSpinner[T]) SetPhases(phases ...string) T {
	s.phases = strutils.AlignLeft(phases, ' ')
	return s.self.Self()
}

func (s *RawSpinner[T]) SetPredefined(i int) T {
	if c, ok := SpinnerTypes[i]; ok {
		s.phases = c
	}
	return s.self.Self()
}

func Visualize[T ProgressInterface[T]](s *RawSpinner[T]) (string, bool) {
	if s.self.Self().IsClosed() {
		return s.done, true
	}
	if !s.self.Self().IsStarted() {
		return "", false
	}
	return s.phases[s.phase], false
}

func (s *RawSpinner[T]) Tick() bool {
	if s.self.Self().IsClosed() {
		return false
	}
	s.self.Self().Start()
	s.lock.Lock()

	s.cnt++
	if s.cnt < s.speed {
		s.lock.Unlock()
		return false
	}
	s.cnt = 0
	s.phase = (s.phase + 1) % len(s.phases)
	s.lock.Unlock()
	return s.self.Protected().Update()
}
