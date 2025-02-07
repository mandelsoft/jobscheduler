package specs

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type BarDef[T any] struct {
	ppi.ProgressBaseDef[T]

	total int

	set    *int
	config *BarConfig
}

// NewBarDef can be used to create a nested definition
// for a derived bar definition.
func NewBarDef[T any](self T) BarDef[T] {
	d := BarDef[T]{total: 100}
	d.ProgressBaseDef = ppi.NewProgressBaseDef[T](self)
	return d

}

func (d *BarDef[T]) Dup(s T) BarDef[T] {
	dup := *d
	dup.ProgressBaseDef = d.ProgressBaseDef.Dup(s)
	return dup
}

func (d *BarDef[T]) SetTotal(v int) T {
	d.total = v
	return d.Self()
}

func (d *BarDef[T]) GetTotal() int {
	return d.total
}

func (d *BarDef[T]) SetPredefined(i int) T {
	d.set = generics.Pointer(i)
	return d.Self()
}

func (d *BarDef[T]) GetPredefined() int {
	if d.set == nil {
		return 0
	}
	return *d.set
}

func (d *BarDef[T]) SetConfig(c BarConfig) T {
	d.config = generics.Pointer(c)
	return d.Self()
}

func (d *BarDef[T]) GetConfig() *BarConfig {
	if d.config == nil {
		return nil
	}
	c := *d.config
	return &c
}
