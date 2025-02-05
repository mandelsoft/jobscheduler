package uiprogress

import (
	"io"

	"github.com/mandelsoft/goutils/general"
	ppi2 "github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

// Text provides a tail window of output
// until the action described by the element is
// calling Text.Close.
// The element can be used as writer to pass the
// intended output.
type Text interface {
	ppi2.BaseInterface

	// SetFinal sets a text message shown instead of the
	// text window after the action has been finished.
	SetFinal(m string) Text

	// SetAuto enables the automatic text window update on
	// calls to Text.Write.
	SetAuto(b ...bool) Text

	// Flush can be called if automatic mode is set to false
	// (the default) to trigger a screen update.
	Flush() error

	io.Writer
}

type _text struct {
	ppi2.ElemBase[Text, ppi2.BaseElement]
}

type _textProtected struct {
	*_text
}

func (t *_textProtected) Update() bool {
	return t._update()
}

// NewText creates a new text stream with the given window size.
// With Text.SetAuto updates are triggered by the Text.Write calls.
// Otherwise, Text.Flush must be called to update the text window.
func NewText(p Progress, view ...int) Text {
	t := &_text{}
	impl := &_textProtected{t}

	t.ElemBase = ppi2.NewElemBase[Text, ppi2.BaseElement](t, impl, p.UIBlocks(), general.OptionalDefaulted(3, view...), nil)
	return t
}

func (t *_text) SetFinal(m string) Text {
	t.UIBlock().SetFinal(m)
	return t
}

func (t *_text) SetAuto(b ...bool) Text {
	t.UIBlock().SetAuto(b...)
	return t
}

func (t *_text) _update() bool {
	return false
}

func (t *_text) Flush() error {
	return t.UIBlock().Flush()
}

func (t *_text) Write(data []byte) (int, error) {
	return t.UIBlock().Write(data)
}
