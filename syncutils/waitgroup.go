package syncutils

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

func NewWaitGroup(h ...utils.WaitingHandler) *WaitGroup {
	return &WaitGroup{waiting: utils.NewWaiting(h...)}
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
