package uiprogress

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type BarDefinition struct {
	specs.BarDef[*BarDefinition]
}

func NewBarDefinition() *BarDefinition {
	d := &BarDefinition{}
	d.BarDef = specs.NewBarDef(d)
	return d
}

func (d *BarDefinition) Dup() *BarDefinition {
	dup := &BarDefinition{}
	dup.BarDef = d.BarDef.Dup(dup)
	return dup
}

func (d *BarDefinition) Add(c Container, total ...int) Bar {
	s := NewBar(c, general.OptionalDefaulted(d.GetTotal(), total...))

	if v := d.GetFinal(); v != "" {
		s.SetFinal(v)
	}
	if v := d.GetColor(); v != nil {
		s.SetColor(v)
	}

	if v := d.GetConfig(); v != nil {
		s.SetBarConfig(*v)
	}

	for _, f := range d.GetAppendFuncs() {
		s.AppendFunc(f)
	}
	for _, f := range d.GetPrependFuncs() {
		s.PrependFunc(f)
	}
	return s
}
