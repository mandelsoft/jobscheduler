package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type TextDefinition struct {
	specs.TextDef[*TextDefinition]
}

func DefineText() *TextDefinition {
	d := &TextDefinition{}
	d.TextDef = specs.NewTextDef(d)
	return d
}

func (d *TextDefinition) Dup() *TextDefinition {
	dup := &TextDefinition{}
	dup.TextDef = d.TextDef.Dup(dup)
	return dup
}

func (d *TextDefinition) Add(c Container) Text {
	s := NewText(c, d.GetView())

	if v := d.GetFinal(); v != "" {
		s.SetFinal(v)
	}
	if v := d.GetFinal(); v != "" {
		s.SetFinal(v)
	}
	if v := d.GetTitleLine(); v != "" {
		s.SetTitleLine(v)
	}

	if d.GetAuto() != nil {
		s.SetAuto(*d.GetAuto())
	}
	return s
}
