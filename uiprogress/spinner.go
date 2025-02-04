package uiprogress

import (
	"maps"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/mandelsoft/jobscheduler/strutils"
)

var CharSets = maps.Clone(spinner.CharSets)

func init() {
	CharSets[1000] = []string{"███▒▒▒▒▒▒▒", "▒███▒▒▒▒▒▒", "▒▒███▒▒▒▒▒", "▒▒▒███▒▒▒▒", "▒▒▒▒███▒▒▒", "▒▒▒▒▒███▒▒", "▒▒▒▒▒▒███▒", "▒▒▒▒▒▒▒███", "█▒▒▒▒▒▒▒██", "██▒▒▒▒▒▒▒█"}
	CharSets[1001] = []string{"███▒▒▒▒▒▒▒", "▒███▒▒▒▒▒▒", "▒▒███▒▒▒▒▒", "▒▒▒███▒▒▒▒", "▒▒▒▒███▒▒▒", "▒▒▒▒▒███▒▒", "▒▒▒▒▒▒███▒", "▒▒▒▒▒▒▒███", "▒▒▒▒▒▒███▒", "▒▒▒▒▒███▒▒", "▒▒▒▒███▒▒▒", "▒▒███▒▒▒▒▒", "▒███▒▒▒▒▒▒"}
	CharSets[1002] = []string{"⋮", "⋰", "⋯", "⋱"}
	CharSets[1003] = []string{"✶", "✷", "✸", "✷"}
	CharSets[1004] = []string{"𝄖", "𝄗", "𝄘", "𝄙", "𝄛", "𝄙", "𝄘", "𝄗", "𝄖"}
}

const DefaultSpeed = 2

type RawSpinner[T ProgressElement] struct {
	lock    sync.Mutex
	self    T
	charset []string
	speed   int

	cnt   int
	phase int
}

func NewRawSpinner[T ProgressElement](self T, set int) RawSpinner[T] {
	if set < 0 || CharSets[set] == nil {
		set = 9
	}
	return RawSpinner[T]{
		self:    self,
		charset: CharSets[set],
		speed:   DefaultSpeed,
	}
}

func (s *RawSpinner[T]) SetSpeed(v int) T {
	s.speed = v
	return s.self
}

func (s *RawSpinner[T]) SetPhases(phases ...string) T {
	s.charset = strutils.AlignLeft(phases, ' ')
	return s.self
}

func (s *RawSpinner[T]) Visualize() (string, bool) {
	if s.self.IsClosed() {
		return "done", true
	}
	return s.charset[s.phase], false
}

func (s *RawSpinner[T]) Tick() bool {
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

type Spinner struct {
	ProgressBase[*Spinner]
	RawSpinner[*Spinner]
	closed bool
}

func NewSpinner(p *Progress, set int) *Spinner {
	b := &Spinner{}
	b.RawSpinner = NewRawSpinner[*Spinner](b, set)
	b.ProgressBase = NewProgressBase[*Spinner](b, p.blocks, ProgressBaseOptions{})
	return b
}

func (s *Spinner) finalize() {
	s.Update()
}

func (s *Spinner) Visualize() (string, bool) {
	if s.self.IsClosed() {
		return "done", true
	}
	return s.RawSpinner.Visualize()
}
