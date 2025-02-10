package syncutils

import (
	"context"
	"sync"

	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

type Monitor interface {
	sync.Locker

	// Wait registers the Go routine in a waiting list and blocks it until
	// the entry is deblocked again by a Signal call.
	// It must be called under a lock, which is released by this call
	// before the go routine is finally blocked.
	// An error is returned, if the blocked go routine is canceled
	// by the context.
	// In any case the lock is locked after Wait returns.
	Wait(ctx context.Context) error

	// Signal deblocks the first waiting go routine, if at least
	// one is blocked. It returns false if no one is found.
	// Signal must be called under the monitor lock.
	// The lock is transferred to the deblocked
	// go routine and true is returned.
	// If no waiting go-routine could be found to transfer the lock to
	// the lock is released and false is returned.
	Signal() bool

	// HasWaiting returns whether there is a blocked Go routine.
	// This method MUST only be called under a monitor lock.
	HasWaiting() bool
}

////////////////////////////////////////////////////////////////////////////////

type monitor struct {
	sync.Locker
	waiting utils.Waiting
}

func NewMonitor(l sync.Locker) Monitor {
	return &monitor{Locker: l}
}

func NewMutexMonitor() Monitor {
	return &monitor{Locker: &sync.Mutex{}}
}

func (w *monitor) Wait(ctx context.Context) error {
	return w.waiting.Wait(ctx, w.Locker)
}

func (w *monitor) Signal() bool {
	return w.waiting.Signal(w.Locker)
}

func (w *monitor) HasWaiting() bool {
	return w.waiting.HasWaiting()
}
