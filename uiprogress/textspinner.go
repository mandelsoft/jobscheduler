package uiprogress

type TextSpinner struct {
	ProgressBase[*TextSpinner]
	RawSpinner[*TextSpinner]
	closed bool
}

func NewTextSpinner(p *Progress, set int, view ...int) *TextSpinner {
	b := &TextSpinner{}
	b.RawSpinner = NewRawSpinner[*TextSpinner](b, set)
	b.ProgressBase = NewProgressBase[*TextSpinner](b, p.blocks, view...)
	b.SetSpeed(4)
	b.block.SetAuto()
	return b
}

func (b *TextSpinner) Write(data []byte) (int, error) {
	return b.block.Write(data)
}

func (s *TextSpinner) Flush() error {
	line, _ := s.Line()
	s.block.SetTitleLine(line)
	return s.block.Flush()
}

func (s *TextSpinner) Close() error {
	if s.RawSpinner.Close() == nil {
		line, _ := s.Line()
		s.block.SetTitleLine(line)
	}
	return s.block.Close()
}
