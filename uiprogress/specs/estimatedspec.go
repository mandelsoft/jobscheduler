package specs

import (
	"time"
)

type EstimatedDef[T any] struct {
	BarBaseDef[T]

	total time.Duration
}

// NewEstimatedDef can be used to create a nested definition
// for a derived bar definition.
func NewEstimatedDef[T any](self T) EstimatedDef[T] {
	d := EstimatedDef[T]{total: 100}
	d.BarBaseDef = NewBarBaseDef[T](self)
	return d

}

func (d *EstimatedDef[T]) Dup(s T) EstimatedDef[T] {
	dup := *d
	dup.BarBaseDef = d.BarBaseDef.Dup(s)
	return dup
}

func (d *EstimatedDef[T]) SetTotal(v time.Duration) T {
	d.total = v
	return d.Self()
}

func (d *EstimatedDef[T]) GetTotal() time.Duration {
	return d.total
}
