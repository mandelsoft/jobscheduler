package processors

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/syncutils"
)

type defaultPool struct {
	monitor   syncutils.Monitor
	available int
}

// NewDefaultPool provides a pool with an implicit capacity.
// every Pool.Alloc will be satisfied, if there was an appropriate
// Pool.Release. Any Go routine can use this pool together with sync
// elements without constraints. No prior reservation is required.
// A new Go routine implicitly increases the pool when calling a blocking
// sync operation. To keep this capacity it should call Release
// before it exits, otherwise the capacity is left unchanged.
func NewDefaultPool(initial ...int) Pool {
	return &defaultPool{monitor: syncutils.NewMutexMonitor(), available: general.Optional(initial...)}
}

func (p *defaultPool) GetPool() Pool {
	return p
}

func (p *defaultPool) Monitor() syncutils.Monitor {
	return p.Monitor()
}

func (p *defaultPool) Release() {
	p.monitor.Lock()
	p.available++
	p.monitor.Signal()
}

func (p *defaultPool) Alloc(ctx context.Context) error {
	p.monitor.Lock()
	defer p.monitor.Unlock()
	if p.available == 0 {
		return p.monitor.Wait(ctx)
	}
	p.available--
	return nil
}
