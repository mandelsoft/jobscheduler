package specs

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type BarBaseDef[T any] struct {
	ppi.ProgressBaseDef[T]
	width   uint
	pending string
	config  BarConfig
}

// NewBarDef can be used to create a nested definition
// for a derived bar definition.
func NewBarBaseDef[T any](self T) BarBaseDef[T] {
	d := BarBaseDef[T]{config: BarTypes[0]}
	d.ProgressBaseDef = ppi.NewProgressBaseDef[T](self)
	return d

}

func (d *BarBaseDef[T]) Dup(s T) BarBaseDef[T] {
	dup := *d
	dup.ProgressBaseDef = d.ProgressBaseDef.Dup(s)
	return dup
}

func (d *BarBaseDef[T]) SetWidth(w uint) T {
	d.width = w
	return d.Self()
}

func (d *BarBaseDef[T]) GetWidth() uint {
	return d.width
}

func (d *BarBaseDef[T]) SetPending(m string) T {
	d.pending = m
	return d.Self()
}

func (d *BarBaseDef[T]) GetPending() string {
	return d.pending
}

func (d *BarBaseDef[T]) SetConfig(c BarConfig) T {
	d.config = c
	return d.Self()
}

func (d *BarBaseDef[T]) GetConfig() BarConfig {
	return d.config
}

func (d *BarBaseDef[T]) SetPredefined(i int) T {
	if c, ok := BarTypes[i]; ok {
		d.config = c
	}
	return d.Self()
}

func (d *BarBaseDef[T]) SetBrackets(c Brackets) T {
	d.config = d.config.SetBackets(c)
	return d.Self()
}

func (d *BarBaseDef[T]) SetPredefinedBrackets(i int) T {
	if c, ok := BracketTypes[i]; ok {
		d.config = d.config.SetBackets(c)
	}
	return d.Self()
}

func (d *BarBaseDef[T]) SetHead(c rune) T {
	d.config.Head = c
	return d.Self()
}

func (d *BarBaseDef[T]) SetEmpty(c rune) T {
	d.config.Empty = c
	return d.Self()
}

func (d *BarBaseDef[T]) SetFill(c rune) T {
	d.config.Fill = c
	return d.Self()
}

func (d *BarBaseDef[T]) SetLeftEnd(c rune) T {
	d.config.LeftEnd = c
	return d.Self()
}

func (d *BarBaseDef[T]) SetRightEnd(c rune) T {
	d.config.RightEnd = c
	return d.Self()
}
