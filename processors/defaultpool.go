package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils"
)

type defaultPool struct {
	monitor   syncutils.Monitor
	available int
}

// NewDefaultPool provides a pool with an implicit capacity.
// every Pool.Alloc will be satisfied, if there was an appropriate
// Pool.Release. Any Go routine can use this pool together with sync
// elements without contraints. No prior reservation is required.
func NewDefaultPool() Pool {
	return &defaultPool{monitor: syncutils.NewMutexMonitor()}
}

func (p *defaultPool) Monitor() syncutils.Monitor {
	return p.Monitor()
}

func (p *defaultPool) Release() {
	p.monitor.Lock()
	p.available++
	if p.available == 1 {
		p.monitor.Signal()
	} else {
		p.monitor.Unlock()
	}
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

////////////////////////////////////////////////////////////////////////////////

// LimitPool is a Pool with a limited initial capacity.
// Any Go routine wanting to use this pool mit call
// a Pool.Alloc first and Pool.Release before finishing.
// It delays the routine until
// there is still capacity. Any blocking sync element using
// this pool with release capacity before blocking and
// alloc it again before continuing.
//
// It limits the parallelity of using Go routines
// to a limit for actually running routines.
type LimitPool interface {
	// Inc increases limit.
	Inc()

	// Dec decreases limit. It waits
	// for avaiable capacity.
	Dec(ctx context.Context) error

	// TryDec decreases the limit, if there
	// is available capacity.
	TryDec() bool

	Limit() int
	Available() int
	Pool
}

type limitpool struct {
	defaultPool
	limit int
}

func NewLimitPool(limit int) LimitPool {
	return &limitpool{defaultPool: defaultPool{monitor: syncutils.NewMutexMonitor(), available: limit}, limit: limit}
}

func (p *limitpool) Monitor() syncutils.Monitor {
	return p.Monitor()
}

func (p *limitpool) Inc() {
	p.monitor.Lock()
	p.limit++
	p.available++
	p.monitor.Signal()
}

func (p *limitpool) Dec(ctx context.Context) error {
	p.monitor.Lock()
	defer p.monitor.Unlock()
	if p.available == 0 {
		return p.monitor.Wait(ctx)
	}
	p.available--
	p.limit--
	return nil
}

func (p *limitpool) TryDec() bool {
	p.monitor.Lock()
	defer p.monitor.Unlock()
	if p.available == 0 {
		return false
	}
	p.available--
	p.limit--
	return true
}

func (p *limitpool) Limit() int {
	p.monitor.Lock()
	defer p.monitor.Unlock()
	return p.limit
}

func (p *limitpool) Available() int {
	p.monitor.Lock()
	defer p.monitor.Unlock()
	return p.available
}

func (p *limitpool) Release() {
	p.monitor.Lock()
	if p.available == p.limit {
		p.monitor.Unlock()
		panic("release exceeds limit")
	}
	p.monitor.Signal()
}
