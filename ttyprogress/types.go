package ttyprogress

import (
	"github.com/mandelsoft/jobscheduler/ttyprogress/ppi"
	"github.com/mandelsoft/jobscheduler/ttyprogress/specs"
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc = specs.DecoratorFunc

type Element = specs.ElementInterface

type Container = ppi.Container

type Ticker interface {
	Tick() bool
}

type ElementSpecification[T Element] interface {
	Add(Container) (T, error)
}
