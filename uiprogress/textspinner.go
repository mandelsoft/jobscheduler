package uiprogress

import (
	"github.com/mandelsoft/goutils/general"
)

type TextSpinner struct {
	ProgressBase[*TextSpinner]
	RawSpinner[*TextSpinner]
	closed bool
}

func NewTextSpinner(p *Progress, set int, view ...int) *TextSpinner {
	b := &TextSpinner{}
	b.RawSpinner = NewRawSpinner[*TextSpinner](b, set)
	b.ProgressBase = NewProgressBase[*TextSpinner](b, p.blocks, ProgressBaseOptions{
		View: general.OptionalDefaulted(3, view...),
	})
	b.SetSpeed(4)
	b.block.SetAuto()
	return b
}

func (b *TextSpinner) Write(data []byte) (int, error) {
	return b.block.Write(data)
}

func (s *TextSpinner) Update() bool {
	line, _ := s.Line()
	s.block.SetTitleLine(line)
	return true
}
