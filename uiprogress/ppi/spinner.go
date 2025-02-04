package ppi

import (
	"maps"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/strutils"
)

var Done = "done"

// SpinnerTypes predefined spinner types.
// Most of them are taken from [spinner.CharSets] (github.com/briandowns/spinner).
var SpinnerTypes = maps.Clone(spinner.CharSets)

func init() {
	SpinnerTypes[1000] = []string{"███▒▒▒▒▒▒▒", "▒███▒▒▒▒▒▒", "▒▒███▒▒▒▒▒", "▒▒▒███▒▒▒▒", "▒▒▒▒███▒▒▒", "▒▒▒▒▒███▒▒", "▒▒▒▒▒▒███▒", "▒▒▒▒▒▒▒███", "█▒▒▒▒▒▒▒██", "██▒▒▒▒▒▒▒█"}
	SpinnerTypes[1001] = []string{"███▒▒▒▒▒▒▒", "▒███▒▒▒▒▒▒", "▒▒███▒▒▒▒▒", "▒▒▒███▒▒▒▒", "▒▒▒▒███▒▒▒", "▒▒▒▒▒███▒▒", "▒▒▒▒▒▒███▒", "▒▒▒▒▒▒▒███", "▒▒▒▒▒▒███▒", "▒▒▒▒▒███▒▒", "▒▒▒▒███▒▒▒", "▒▒███▒▒▒▒▒", "▒███▒▒▒▒▒▒"}
	SpinnerTypes[1002] = []string{"⋮", "⋰", "⋯", "⋱"}
	SpinnerTypes[1003] = []string{"✶", "✷", "✸", "✷"}
	SpinnerTypes[1004] = []string{"𝄖", "𝄗", "𝄘", "𝄙", "𝄛", "𝄙", "𝄘", "𝄗", "𝄖"}
}

var Speed = 2

type RawSpinner[P ProgressInterface[P], T ProgressImplementation[P]] struct {
	lock    sync.Mutex
	self    T
	charset []string
	speed   int
	done    string

	cnt   int
	phase int
}

func NewRawSpinner[P ProgressInterface[P], T ProgressImplementation[P]](self T, set int) RawSpinner[P, T] {
	if set < 0 || SpinnerTypes[set] == nil {
		set = 9
	}
	return RawSpinner[P, T]{
		self:    self,
		charset: SpinnerTypes[set],
		speed:   Speed,
		done:    Done,
	}
}

func (s *RawSpinner[P, T]) SetSpeed(v int) P {
	s.speed = v
	return generics.Cast[P](s.self)
}

func (s *RawSpinner[P, T]) SetDone(m string) P {
	s.done = m
	return generics.Cast[P](s.self)
}

func (s *RawSpinner[P, T]) SetPhases(phases ...string) P {
	s.charset = strutils.AlignLeft(phases, ' ')
	return generics.Cast[P](s.self)
}

func (s *RawSpinner[P, T]) Visualize() (string, bool) {
	if s.self.IsClosed() {
		return s.done, true
	}
	if !s.self.IsStarted() {
		return "", false
	}
	return s.charset[s.phase], false
}

func (s *RawSpinner[P, T]) Tick() bool {
	if s.self.IsClosed() {
		return false
	}
	s.self.Start()
	s.lock.Lock()

	s.cnt++
	if s.cnt < s.speed {
		s.lock.Unlock()
		return false
	}
	s.cnt = 0
	s.phase = (s.phase + 1) % len(s.charset)
	s.lock.Unlock()
	return s.self.Update()
}

////////////////////////////////////////////////////////////////////////////////

// Spinner provides one line of unlimited progress information.
type Spinner interface {
	ProgressInterface[Spinner]
	SetSpeed(int) Spinner
}

type spinnerImpl struct {
	ProgressBase[Spinner, *spinnerImpl]
	RawSpinner[Spinner, *spinnerImpl]
	closed bool
}

// NewSpinner creates a Spinner with a predefined
// set of spinner phases taken from SpinnerTypes.
func NewSpinner(p Progress, set int) Spinner {
	b := &spinnerImpl{}
	b.RawSpinner = NewRawSpinner[Spinner](b, set)
	b.ProgressBase = NewProgressBase[Spinner](b, p.UIBlocks(), 1, nil)
	return b
}

func (s *spinnerImpl) finalize() {
	s.Update()
}

func (s *spinnerImpl) Visualize() (string, bool) {
	if s.self.IsClosed() {
		return "done", true
	}
	return s.RawSpinner.Visualize()
}
