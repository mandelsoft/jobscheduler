package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type StepsDefinition struct {
	specs.StepsDef[*StepsDefinition]
}

func DefineSteps(steps ...string) *StepsDefinition {
	d := &StepsDefinition{}
	d.StepsDef = specs.NewStepsDef(d, steps)
	return d
}

func (d *StepsDefinition) Dup() *StepsDefinition {
	dup := &StepsDefinition{}
	dup.StepsDef = d.StepsDef.Dup(dup)
	return dup
}

func (d *StepsDefinition) Add(c Container) Bar {
	s := NewSteps(c, d.GetSteps())

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
