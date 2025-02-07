package specs

import (
	"slices"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type StepsDef[T any] struct {
	ppi.ProgressBaseDef[T]

	steps []string

	set    *int
	config *BarConfig
}

// NewStepsDef can be used to create a nested definition
// for a derived steps definition.
func NewStepsDef[T any](self T, steps []string) StepsDef[T] {
	d := StepsDef[T]{steps: slices.Clone(steps)}
	d.ProgressBaseDef = ppi.NewProgressBaseDef[T](self)
	return d

}

func (d *StepsDef[T]) Dup(s T) StepsDef[T] {
	dup := *d
	dup.ProgressBaseDef = d.ProgressBaseDef.Dup(s)
	return dup
}

func (d *StepsDef[T]) GetSteps() []string {
	return slices.Clone(d.steps)
}

func (d *StepsDef[T]) SetPredefined(i int) T {
	d.set = generics.Pointer(i)
	return d.Self()
}

func (d *StepsDef[T]) GetPredefined() int {
	if d.set == nil {
		return 0
	}
	return *d.set
}

func (d *StepsDef[T]) SetConfig(c BarConfig) T {
	d.config = generics.Pointer(c)
	return d.Self()
}

func (d *StepsDef[T]) GetConfig() *BarConfig {
	if d.config == nil {
		return nil
	}
	c := *d.config
	return &c
}
