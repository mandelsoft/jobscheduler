package specs

type TextSpinnerDef[T any] struct {
	SpinnerDef[T]

	view int
	gap  string
}

// NewTextSpinnerDef can be used to create a nested definition
// for a derived text spinner definition.
func NewTextSpinnerDef[T any](self T) TextSpinnerDef[T] {
	d := TextSpinnerDef[T]{view: 3}
	d.SpinnerDef = NewSpinnerDef[T](self)
	return d
}

func (d *TextSpinnerDef[T]) Dup(s T) TextSpinnerDef[T] {
	dup := *d
	dup.SpinnerDef = d.SpinnerDef.Dup(s)
	return dup
}

func (d *TextSpinnerDef[T]) SetView(view int) T {
	d.view = view
	return d.Self()
}

func (d *TextSpinnerDef[T]) GetView() int {
	return d.view
}

func (d *TextSpinnerDef[T]) SetGap(gap string) T {
	d.gap = gap
	return d.Self()
}

func (d *TextSpinnerDef[T]) GetGap() string {
	return d.gap
}
