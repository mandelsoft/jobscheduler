package uiprogress

import (
	"maps"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
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
	SpinnerTypes[1005] = []string{"▢", "▣"}
	SpinnerTypes[1006] = []string{"◇", "◈"}
	SpinnerTypes[1007] = []string{"◇", "◈", "◆"}
}

var Speed = 2

type RawSpinner[P ProgressInterface[P]] struct {
	lock sync.Mutex
	self ppi.Self[P, ppi.ProgressProtected[P]]

	charset []string
	speed   int
	done    string

	cnt   int
	phase int
}

func NewRawSpinner[T ProgressInterface[T]](self ppi.Self[T, ppi.ProgressProtected[T]], set int) RawSpinner[T] {
	if set < 0 || SpinnerTypes[set] == nil {
		set = 9
	}
	return RawSpinner[T]{
		self:    self,
		charset: SpinnerTypes[set],
		cnt:     Speed - 1,
		speed:   Speed,
		done:    Done,
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
	s.charset = strutils.AlignLeft(phases, ' ')
	return s.self.Self()
}

func Visualize[T ProgressInterface[T]](s *RawSpinner[T]) (string, bool) {
	if s.self.Self().IsClosed() {
		return s.done, true
	}
	if !s.self.Self().IsStarted() {
		return "", false
	}
	return s.charset[s.phase], false
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
	s.phase = (s.phase + 1) % len(s.charset)
	s.lock.Unlock()
	return s.self.Protected().Update()
}
