package uiprogress

import (
	"io"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

// TextSpinner is a combination of a Spinner and
// a Text element. A title line reports the progress
// followed by a text window.
type TextSpinner interface {
	ppi.ProgressInterface[TextSpinner]
	RawSpinnerInterface[TextSpinner]

	// SetAuto enables the automatic text window update on
	// calls to Text.Write. Even if not set, the update
	// is triggered by every spinner update.
	SetAuto(b ...bool) TextSpinner

	// SetGap sets the relative gap for the content.
	SetGap(gap string) TextSpinner

	io.Writer
}

type _TextSpinner struct {
	ppi.ProgressBase[TextSpinner]
	RawSpinner[TextSpinner]
	closed bool
}

type _textSpinnerProtected struct {
	*_TextSpinner
}

func (t *_textSpinnerProtected) Self() TextSpinner {
	return t._TextSpinner
}

func (t *_textSpinnerProtected) Update() bool {
	return t._update()
}

func (t *_textSpinnerProtected) Visualize() (string, bool) {
	return t._visualize()
}

// NewTextSpinner creates  new TextSpinner with the given
// window size. It can be used like a Text element.
func NewTextSpinner(p Container, set int, view ...int) TextSpinner {
	b := &_TextSpinner{}
	self := ppi.ProgressSelf[TextSpinner](&_textSpinnerProtected{b})

	b.RawSpinner = NewRawSpinner[TextSpinner](self, set)
	b.ProgressBase = ppi.NewProgressBase[TextSpinner](self, p, general.OptionalDefaulted(3, view...), nil)
	b.SetSpeed(4)
	b.UIBlock().SetAuto()
	return b
}

func (t *_TextSpinner) SetAuto(b ...bool) TextSpinner {
	t.UIBlock().SetAuto(b...)
	return t
}

func (t *_TextSpinner) SetGap(gap string) TextSpinner {
	t.UIBlock().SetContentGap(gap)
	return t
}

func (b *_TextSpinner) Write(data []byte) (int, error) {
	b.Start()
	return b.UIBlock().Write(data)
}

func (s *_TextSpinner) _update() bool {
	line, _ := s.Line()
	s.UIBlock().SetTitleLine(line)
	return true
}

func (s *_TextSpinner) _visualize() (string, bool) {
	return Visualize(&s.RawSpinner)
}
