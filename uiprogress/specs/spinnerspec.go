package specs

import (
	"slices"

	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type SpinnerDef[T any] struct {
	ppi.ProgressBaseDef[T]

	set    *int
	done   string
	speed  int
	phases []string
}

// NewSpinnerDef can be used to create a nested definition
// for a derived spinner definition.
func NewSpinnerDef[T any](self T) SpinnerDef[T] {
	d := SpinnerDef[T]{speed: 1}
	d.ProgressBaseDef = ppi.NewProgressBaseDef[T](self)
	return d

}

func (d *SpinnerDef[T]) Dup(s T) SpinnerDef[T] {
	dup := *d
	dup.ProgressBaseDef = d.ProgressBaseDef.Dup(s)
	return dup
}

func (d *SpinnerDef[T]) SetPredefined(i int) T {
	if c, ok := SpinnerTypes[i]; ok {
		d.phases = c
	}
	return d.Self()
}

func (d *SpinnerDef[T]) SetDone(m string) T {
	d.done = m
	return d.Self()
}

func (d *SpinnerDef[T]) GetDone() string {
	return d.done
}

func (d *SpinnerDef[T]) SetSpeed(v int) T {
	d.speed = v
	return d.Self()
}

func (d *SpinnerDef[T]) GetSpeed() int {
	return d.speed
}

func (d *SpinnerDef[T]) SetPhases(p []string) T {
	d.phases = slices.Clone(p)
	return d.Self()
}

func (d *SpinnerDef[T]) GetPhases() []string {
	return slices.Clone(d.phases)
}
