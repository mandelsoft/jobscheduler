package uiprogress

import (
	"time"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/stringutils"
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

type EstimatedDefinition struct {
	specs.EstimatedDef[*EstimatedDefinition]
}

func NewEstimatedDefinition() *EstimatedDefinition {
	d := &EstimatedDefinition{}
	d.EstimatedDef = specs.NewEstimatedDef(d)
	return d
}

// PrependEstimated prepends the time elapsed to the begining of the bar
func (b *EstimatedDefinition) PrependEstimated(offset ...int) *EstimatedDefinition {
	b.PrependFunc(func(e Element) string {
		return stringutils.PadLeft(e.(Estimated).TimeEstimatedString(), 5, ' ')
	}, offset...)
	return b
}

// AppendEstimated appends the time elapsed to the begining of the bar
func (b *EstimatedDefinition) AppendEstimated(offset ...int) *EstimatedDefinition {
	b.AppendFunc(func(e Element) string {
		return stringutils.PadLeft(e.(Estimated).TimeEstimatedString(), 5, ' ')
	}, offset...)
	return b
}

func (d *EstimatedDefinition) Dup() *EstimatedDefinition {
	dup := &EstimatedDefinition{}
	dup.EstimatedDef = d.EstimatedDef.Dup(dup)
	return dup
}

func (d *EstimatedDefinition) Add(c Container, total ...time.Duration) Estimated {
	s := NewEstimated(c, general.OptionalDefaulted(d.GetTotal(), total...))

	if v := d.GetFinal(); v != "" {
		s.SetFinal(v)
	}
	if v := d.GetColor(); v != nil {
		s.SetColor(v)
	}
	if v := d.GetPending(); v != "" {
		s.SetPending(v)

	}
	if v := d.GetWidth(); v != 0 {
		s.SetWidth(v)
	}

	s.SetBarConfig(d.GetConfig())

	for _, f := range d.GetAppendFuncs() {
		s.AppendFunc(f)
	}
	for _, f := range d.GetPrependFuncs() {
		s.PrependFunc(f)
	}
	return s
}
