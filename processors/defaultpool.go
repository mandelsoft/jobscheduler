package processors

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/syncutils"
)

type defaultPool struct {
	monitor   syncutils.Monitor
	available int
	self      Pool
}

// NewDefaultPool provides a pool with an implicit capacity.
// every Pool.Alloc will be satisfied, if there was an appropriate
// Pool.Release. Any Go routine can use this pool together with sync
// elements without constraints. No prior reservation is required.
// A new Go routine implicitly increases the pool when calling a blocking
// sync operation. To keep this capacity it should call Release
// before it exits, otherwise the capacity is left unchanged.
func NewDefaultPool(initial ...int) Pool {
	p := &defaultPool{monitor: syncutils.NewMutexMonitor(), available: general.Optional(initial...)}
	p.self = p
	return p
}

func (p *defaultPool) GetPool() Pool {
	return p.self
}

func (p *defaultPool) Monitor() syncutils.Monitor {
	return p.Monitor()
}

func (p *defaultPool) Release(ctx context.Context) {
	p.monitor.Lock()
	p.available++
	p.monitor.Signal(ctx)
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
