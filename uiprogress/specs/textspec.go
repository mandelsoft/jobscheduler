package specs

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type TextDef[T any] struct {
	ppi.ElemBaseDef[T]
	view int

	auto      *bool
	titleline string
}

// NewTextDef can be used to create a nested definition
// for a derived text definition.
func NewTextDef[T any](self T) TextDef[T] {
	d := TextDef[T]{view: 3}
	d.ElemBaseDef = ppi.NewBaseDef[T](self)
	return d
}

func (d *TextDef[T]) Dup(s T) TextDef[T] {
	dup := *d
	dup.ElemBaseDef = d.ElemBaseDef.Dup(s)
	return dup
}

func (d *TextDef[T]) GetView() int {
	return d.view
}

func (d *TextDef[T]) GetAuto() *bool {
	return d.auto
}

func (d *TextDef[T]) SetAuto(b ...bool) T {
	d.auto = generics.Pointer(general.OptionalDefaultedBool(true, b...))
	return d.Self()
}

func (d *TextDef[T]) SetTitleLine(v string) T {
	d.titleline = v
	return d.Self()
}

func (d *TextDef[T]) GetTitleLine() string {
	return d.titleline
}
