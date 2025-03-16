package specs

import (
	"slices"

	"github.com/fatih/color"
	"github.com/mandelsoft/goutils/stringutils"
)

type ProgressInterface interface {
	ElementInterface
}

////////////////////////////////////////////////////////////////////////////////

type ProgressDefinition[T any] struct {
	ElementDefinition[T]

	color        *color.Color
	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc
	tick         bool
}

var (
	_ ProgressSpecification[any] = (*ProgressDefinition[any])(nil)
	_ ProgressConfiguration      = (*ProgressDefinition[any])(nil)
)

func NewProgressDefinition[T any](self Self[T]) ProgressDefinition[T] {
	return ProgressDefinition[T]{ElementDefinition: NewElementDefinition(self)}
}

func (d *ProgressDefinition[T]) Dup(s Self[T]) ProgressDefinition[T] {
	dup := *d
	dup.ElementDefinition = d.ElementDefinition.Dup(s)
	dup.appendFuncs = slices.Clone(dup.appendFuncs)
	dup.prependFuncs = slices.Clone(dup.prependFuncs)
	return dup
}

// GetTick returns whether a tick is required.
func (d *ProgressDefinition[T]) GetTick() bool {
	return d.tick
}

// SetColor appends the time elapsed the be progress bar
func (d *ProgressDefinition[T]) SetColor(col *color.Color) T {
	d.color = col
	return d.Self()
}

func (d *ProgressDefinition[T]) GetColor() *color.Color {
	return d.color
}

// AppendFunc runs the decorator function and renders the output on the right of the progress bar
func (d *ProgressDefinition[T]) AppendFunc(f DecoratorFunc, offset ...int) T {
	if len(offset) == 0 {
		d.appendFuncs = append(d.appendFuncs, f)
	} else {
		d.appendFuncs = slices.Insert(d.appendFuncs, offset[0], f)
	}
	return d.Self()
}

func (d *ProgressDefinition[T]) GetAppendFuncs() []DecoratorFunc {
	return slices.Clone(d.appendFuncs)
}

// PrependFunc runs decorator function and render the output left the progress bar
func (d *ProgressDefinition[T]) PrependFunc(f DecoratorFunc, offset ...int) T {
	if len(offset) == 0 {
		d.prependFuncs = append(d.prependFuncs, f)
	} else {
		d.prependFuncs = slices.Insert(d.prependFuncs, offset[0], f)
	}
	return d.Self()
}

func (d *ProgressDefinition[T]) GetPrependFuncs() []DecoratorFunc {
	return slices.Clone(d.prependFuncs)
}

// AppendElapsed appends the time elapsed the be progress bar
func (d *ProgressDefinition[T]) AppendElapsed(offset ...int) T {
	d.tick = true
	return d.AppendFunc(func(e ElementInterface) string {
		return stringutils.PadLeft(e.TimeElapsedString(), 5, ' ')
	}, offset...)
}

// PrependElapsed prepends the time elapsed to the beginning of the bar
func (d *ProgressDefinition[T]) PrependElapsed(offset ...int) T {
	d.tick = true
	return d.PrependFunc(func(e ElementInterface) string {
		return stringutils.PadLeft(e.TimeElapsedString(), 5, ' ')
	}, offset...)
}

////////////////////////////////////////////////////////////////////////////////

// ProgressSpecification is the configuration interface for progress indicators.
type ProgressSpecification[T any] interface {
	ElementSpecification[T]

	// SetColor set the color used for the progress line.
	SetColor(col *color.Color) T

	// AppendFunc adds a function providing some text appended
	// to the basic progress indicator.
	// If there are implicit settings, the offset can be used to
	// specify the index in the list of functions.
	AppendFunc(f DecoratorFunc, offset ...int) T

	// PrependFunc adds a function providing some text prepended
	// to the basic progress indicator.
	// If there are implicit settings, the offset can be used to
	// specify the index in the list of functions.
	PrependFunc(f DecoratorFunc, offset ...int) T

	// AppendElapsed appends the elapsed time of the action
	// or the duration of the action if the element is already closed.
	AppendElapsed(offset ...int) T

	// PrependElapsed appends the elapsed time of the action
	// or the duration of the action if the element is already closed.
	PrependElapsed(offset ...int) T
}

// ProgressConfiguration provides access to the generic Progress configuration
type ProgressConfiguration interface {
	ElementConfiguration
	GetTick() bool

	GetColor() *color.Color
	GetPrependFuncs() []DecoratorFunc
	GetAppendFuncs() []DecoratorFunc
}
