package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type SpinnerDefinition struct {
	specs.SpinnerDef[*SpinnerDefinition]
}

func DefineSpinner() *SpinnerDefinition {
	d := &SpinnerDefinition{}
	d.SpinnerDef = specs.NewSpinnerDef(d)
	return d
}

func (d *SpinnerDefinition) Dup() *SpinnerDefinition {
	dup := &SpinnerDefinition{}
	dup.SpinnerDef = d.SpinnerDef.Dup(dup)
	return dup
}

func (d *SpinnerDefinition) Add(c Container) Spinner {
	s := NewSpinner(c, 0)

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
