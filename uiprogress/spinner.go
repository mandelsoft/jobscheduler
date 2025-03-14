package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

var SpinnerTypes = specs.SpinnerTypes

// Spinner provides one line of unlimited progress information.
type Spinner interface {
	RawSpinnerInterface[Spinner]
}

type _Spinner struct {
	RawSpinner[Spinner]
	closed bool
}

type _spinnerProtected struct {
	*_Spinner
}

func (s *_spinnerProtected) Self() Spinner {
	return s._Spinner
}

func (s *_spinnerProtected) Update() bool {
	return s._update()
}

func (s *_spinnerProtected) Visualize() (string, bool) {
	return s._visualize()
}

// NewSpinner creates a Spinner with a predefined
// set of spinner phases taken from SpinnerTypes.
func NewSpinner(p Container, set int) Spinner {
	b := &_Spinner{}

	self := ppi.ProgressSelf[Spinner](&_spinnerProtected{b})
	b.RawSpinner = NewRawSpinner[Spinner](self, set, p, 1, nil)
	return b
}

func (s *_Spinner) finalize() {
	s._update()
}

func (s *_Spinner) _update() bool {
	return ppi.Update(&s.ProgressBase)
}

func (s *_Spinner) _visualize() (string, bool) {
	if s.self.Self().IsClosed() {
		return "done", true
	}
	return Visualize(&s.RawSpinner)
}
