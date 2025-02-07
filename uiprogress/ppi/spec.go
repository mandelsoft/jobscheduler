package ppi

import (
	"slices"

	"github.com/fatih/color"
	"github.com/mandelsoft/jobscheduler/strutils"
)

type ElemBaseDef[T any] struct {
	self  T
	final string
}

func NewBaseDef[T any](self T) ElemBaseDef[T] {
	return ElemBaseDef[T]{self: self}
}

func (d *ElemBaseDef[T]) Self() T {
	return d.self
}

func (d *ElemBaseDef[T]) Dup(s T) ElemBaseDef[T] {
	dup := *d
	dup.self = s
	return dup
}

func (d *ElemBaseDef[T]) SetFinal(final string) T {
	d.final = final
	return d.self
}

func (d *ElemBaseDef[T]) GetFinal() string {
	return d.final
}

////////////////////////////////////////////////////////////////////////////////

type ProgressBaseDef[T any] struct {
	ElemBaseDef[T]

	color *color.Color

	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc
}

func NewProgressBaseDef[T any](self T) ProgressBaseDef[T] {
	return ProgressBaseDef[T]{ElemBaseDef: NewBaseDef(self)}
}

func (d *ProgressBaseDef[T]) Dup(s T) ProgressBaseDef[T] {
	dup := *d
	dup.ElemBaseDef = d.ElemBaseDef.Dup(s)
	dup.appendFuncs = slices.Clone(dup.appendFuncs)
	dup.prependFuncs = slices.Clone(dup.prependFuncs)
	return dup
}

// SetColor appends the time elapsed the be progress bar
func (d *ProgressBaseDef[T]) SetColor(col *color.Color) T {
	d.color = col
	return d.Self()
}

func (d *ProgressBaseDef[T]) GetColor() *color.Color {
	return d.color
}

// AppendFunc runs the decorator function and renders the output on the right of the progress bar
func (d *ProgressBaseDef[T]) AppendFunc(f DecoratorFunc, offset ...int) T {
	if len(offset) == 0 {
		d.appendFuncs = append(d.appendFuncs, f)
	} else {
		d.appendFuncs = slices.Insert(d.appendFuncs, offset[0], f)
	}
	return d.Self()
}

func (d *ProgressBaseDef[T]) GetAppendFuncs() []DecoratorFunc {
	return slices.Clone(d.appendFuncs)
}

// PrependFunc runs decorator function and render the output left the progress bar
func (d *ProgressBaseDef[T]) PrependFunc(f DecoratorFunc, offset ...int) T {
	if len(offset) == 0 {
		d.prependFuncs = append(d.prependFuncs, f)
	} else {
		d.prependFuncs = slices.Insert(d.prependFuncs, offset[0], f)
	}
	return d.Self()
}

func (d *ProgressBaseDef[T]) GetPrependFuncs() []DecoratorFunc {
	return slices.Clone(d.prependFuncs)
}

// AppendElapsed appends the time elapsed the be progress bar
func (d *ProgressBaseDef[T]) AppendElapsed(offset ...int) T {
	return d.AppendFunc(func(e Element) string {
		return strutils.PadLeft(e.(BaseInterface).TimeElapsedString(), 5, ' ')
	}, offset...)
}

// PrependElapsed prepends the time elapsed to the beginning of the bar
func (d *ProgressBaseDef[T]) PrependElapsed(offset ...int) T {
	return d.PrependFunc(func(e Element) string {
		return strutils.PadLeft(e.(BaseInterface).TimeElapsedString(), 5, ' ')
	}, offset...)
}
