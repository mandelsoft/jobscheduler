package ppi

import (
	"io"

	"github.com/mandelsoft/goutils/general"
)

// Text provides a tail window of output
// until the action described by the element is
// calling Text.Close.
// The element can be used as writer to pass the
// intended output.
type Text interface {
	BaseInterface

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

type textImpl struct {
	ElemBase[*textImpl]
}

// NewText creates a new text stream with the given window size.
// With Text.SetAuto updates are triggered by the Text.Write calls.
// Otherwise, Text.Flush must be called to update the text window.
func NewText(p Progress, view ...int) Text {
	t := &textImpl{}
	t.ElemBase = NewElemBase(t, p.UIBlocks(), general.OptionalDefaulted(3, view...), nil)
	return t
}

func (t *textImpl) SetFinal(m string) Text {
	t.block.SetFinal(m)
	return t
}

func (t *textImpl) SetAuto(b ...bool) Text {
	t.block.SetAuto(b...)
	return t
}

func (t *textImpl) Update() bool {
	return false
}

func (t *textImpl) Flush() error {
	return t.block.Flush()
}

func (t *textImpl) Write(data []byte) (int, error) {
	return t.block.Write(data)
}
