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

func (wg *WaitGroup) Wait(ctx context.Context) {
	wg.lock.Lock()
	if wg.count == 0 {
		wg.lock.Unlock()
		return
	}
	wg.waiting.Wait(ctx, &wg.lock)
}
