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
	Signal(ctx context.Context) bool

	// HasWaiting returns whether there is a blocked Go routine.
	// This method MUST only be called under a monitor lock.
	HasWaiting() bool

	// SignalAll deblocks all waiting Go routines.
	// It must be called under the monitor lock and does
	// not transfer the lock to any deblocked Go routine.
	// Each routine acquires a lock separately.
	SignalAll() bool
}

////////////////////////////////////////////////////////////////////////////////

type monitor struct {
	sync.Locker
	waiting utils.Waiting
}

func NewMonitor(l sync.Locker, h ...utils.WaitingHandler) Monitor {
	return &monitor{Locker: l, waiting: utils.NewWaiting(h...)}
}

func NewMutexMonitor(h ...utils.WaitingHandler) Monitor {
	return NewMonitor(&sync.Mutex{}, h...)
}

func (w *monitor) Wait(ctx context.Context) error {
	return w.waiting.Wait(ctx, w.Locker)
}

func (w *monitor) Signal(ctx context.Context) bool {
	return w.waiting.Signal(ctx, w.Locker)
}

func (w *monitor) SignalAll() bool {
	return w.waiting.SignalAll()
}

func (w *monitor) HasWaiting() bool {
	return w.waiting.HasWaiting()
}
