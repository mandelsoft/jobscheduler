package processors

import (
	"context"
	"sync"

	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

type WaitGroup struct {
	lock    sync.Mutex
	count   int
	waiting utils.Waiting
}

// NewWaitGroup creates a new WaitGroup working on
// a Pool. The pool must be bound to the context.Context.
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{waiting: utils.NewWaiting(&limithandler{})}
}

func (wg *WaitGroup) Add(delta int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	if wg.count+delta < 0 {
		panic("negative count")
	}
	wg.count += delta
	if wg.count == 0 {
		wg.waiting.SignalAll()
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait(ctx context.Context) error {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	if wg.count == 0 {
		return nil
	}
	return wg.waiting.Wait(ctx, &wg.lock)
}
