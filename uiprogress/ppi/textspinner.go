package ppi

import (
	"io"

	"github.com/mandelsoft/goutils/general"
)

// TextSpinner is a combination of a Spinner and
// a Text element. A title line reports the progress
// followed by a text window.
type TextSpinner interface {
	ProgressInterface[TextSpinner]

	// SetAuto enables the automatic text window update on
	// calls to Text.Write. Even if not set, the update
	// is triggered by every spinner update.
	SetAuto(b ...bool) TextSpinner

	io.Writer
}

type textSpinnerImpl struct {
	ProgressBase[TextSpinner, *textSpinnerImpl]
	RawSpinner[TextSpinner, *textSpinnerImpl]
	closed bool
}

// NewTextSpinner creates  new TextSpinner with the given
// window size. It can be used like a Text element.
func NewTextSpinner(p Progress, set int, view ...int) TextSpinner {
	b := &textSpinnerImpl{}
	b.RawSpinner = NewRawSpinner[TextSpinner](b, set)
	b.ProgressBase = NewProgressBase[TextSpinner](b, p.UIBlocks(), general.OptionalDefaulted(3, view...), nil)
	b.SetSpeed(4)
	b.block.SetAuto()
	return b
}

func (t *textSpinnerImpl) SetAuto(b ...bool) TextSpinner {
	t.block.SetAuto(b...)
	return t
}

func (b *textSpinnerImpl) Write(data []byte) (int, error) {
	return b.block.Write(data)
}

func (s *textSpinnerImpl) Update() bool {
	line, _ := s.Line()
	s.block.SetTitleLine(line)
	return true
}
