package specs

type BarDef[T any] struct {
	BarBaseDef[T]

	total int
}

// NewBarDef can be used to create a nested definition
// for a derived bar definition.
func NewBarDef[T any](self T) BarDef[T] {
	d := BarDef[T]{total: 100}
	d.BarBaseDef = NewBarBaseDef[T](self)
	return d
}

func (d *BarDef[T]) Dup(s T) BarDef[T] {
	dup := *d
	dup.BarBaseDef = d.BarBaseDef.Dup(s)
	return dup
}

func (d *BarDef[T]) SetTotal(v int) T {
	d.total = v
	return d.Self()
}

func (d *BarDef[T]) GetTotal() int {
	return d.total
}
