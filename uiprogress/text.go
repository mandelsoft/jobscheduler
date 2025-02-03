package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiblocks"
)

type Text struct {
	block *uiblocks.Block
}

func NewText(p *Progress, view ...int) *Text {
	return &Text{p.blocks.NewBlock(view...)}
}

func (t *Text) SetFinal(m string) *Text {
	t.block.SetFinal(m)
	return t
}
