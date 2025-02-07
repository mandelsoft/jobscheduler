package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type TextSpinnerDefinition struct {
	specs.TextSpinnerDef[*TextSpinnerDefinition]
}

func DefineTextSpinner() *TextSpinnerDefinition {
	d := &TextSpinnerDefinition{}
	d.TextSpinnerDef = specs.NewTextSpinnerDef(d)
	return d
}

func (d *TextSpinnerDefinition) Dup() *TextSpinnerDefinition {
	dup := &TextSpinnerDefinition{}
	dup.TextSpinnerDef = d.TextSpinnerDef.Dup(dup)
	return dup
}

func (d *TextSpinnerDefinition) Add(c Container) TextSpinner {
	s := NewTextSpinner(c, 0, d.GetView())

	if v := d.GetGap(); v != "" {
		s.SetGap(v)
	}
	if v := d.GetFinal(); v != "" {
		s.SetFinal(v)
	}
	if v := d.GetColor(); v != nil {
		s.SetColor(v)
	}
	if v := d.GetSpeed(); v != 0 {
		s.SetSpeed(v)
	}
	if v := d.GetPhases(); v != nil {
		s.SetPhases(v...)
	}

	for _, f := range d.GetAppendFuncs() {
		s.AppendFunc(f)
	}
	for _, f := range d.GetPrependFuncs() {
		s.PrependFunc(f)
	}

	return s
}
