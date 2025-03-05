package processors

import (
	"context"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

type WaitGroup struct {
	lock    sync.Mutex
	pool    Pool
	count   int
	waiting utils.Waiting
}

func NewWaitGroup(pool Pool) *WaitGroup {
	return &WaitGroup{pool: pool}
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
	wg.pool.Release()
	err := wg.waiting.Wait(ctx, &wg.lock)
	err2 := wg.pool.Alloc(ctx)
	return errors.Join(err, err2)
}
