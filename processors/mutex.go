package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils"
)

var _ context.Context = nil

type Locker interface {
	syncutils.Locker
}

type Mutex interface {
	Locker
}

// NewMutex creates a new Mutex working on
// a Pool. The pool must be bound to the context.Context.
func NewMutex() Mutex {
	return syncutils.NewMutex2(NewMonitor())
}
