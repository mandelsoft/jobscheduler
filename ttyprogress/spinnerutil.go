package ttyprogress

import (
	"sync"

	"github.com/mandelsoft/jobscheduler/ttyprogress/ppi"
	"github.com/mandelsoft/jobscheduler/ttyprogress/specs"
)

var SpinnerTypes = specs.SpinnerTypes

type RawSpinnerInterface interface {
	ppi.ProgressInterface
}

type RawSpinner[P ppi.ProgressInterface] struct {
	ppi.ProgressBase[P]

	lock sync.Mutex
	self ppi.Self[P, ppi.ProgressProtected[P]]

	// pending is the message shown before started
	pending string

	// done is the message shown after closed
	done string

	phases []string
	speed  int

	cnt   int
	phase int
}

var _ RawSpinnerInterface = (*RawSpinner[ppi.ProgressInterface])(nil)

func NewRawSpinner[T ppi.ProgressInterface](self ppi.Self[T, ppi.ProgressProtected[T]], p Container, c specs.SpinnerConfiguration, view int, closer func()) (*RawSpinner[T], error) {
	e := &RawSpinner[T]{
		self:    self,
		phases:  c.GetPhases(),
		cnt:     c.GetSpeed() - 1,
		speed:   c.GetSpeed(),
		done:    c.GetDone(),
		pending: c.GetPending(),
	}
	b, err := ppi.NewProgressBase[T](self, p, c, view, closer, true)
	if err != nil {
		return nil, err
	}
	e.ProgressBase = *b
	return e, nil
}

func (s *RawSpinner[T]) SetSpeed(v int) T {
	s.speed = v
	s.cnt = v - 1
	return s.self.Self()
}

func Visualize[T ppi.ProgressInterface](s *RawSpinner[T]) (string, bool) {
	if s.self.Self().IsClosed() {
		return s.done, true
	}
	if !s.self.Self().IsStarted() {
		return s.pending, false
	}
	return s.phases[s.phase], false
}

func (s *RawSpinner[T]) Tick() bool {
	if s.self.Self().IsClosed() {
		return false
	}
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
