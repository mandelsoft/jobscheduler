package processors

import (
	"context"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/jobscheduler/syncutils"
)

// ErrPoolEmpty is returned if an already empty pools should
// be decreased.
var ErrPoolEmpty = errors.New("pool is empty")
var ErrAlreadyStarted = errors.New("pool is already started")

////////////////////////////////////////////////////////////////////////////////

// LimitPool is a Pool with a limited initial capacity.
// Any Go routine wanting to use this pool must call
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
	p := &limitpool{defaultPool: defaultPool{monitor: syncutils.NewMutexMonitor(), available: limit}, limit: limit}
	p.self = p
	return p
}

func (p *limitpool) Inc() {
	p.monitor.Lock()
	p.limit++
	p.available++
	p.monitor.Signal(nil)
}

func (p *limitpool) Dec(ctx context.Context) error {
	p.monitor.Lock()
	if p.limit == 0 {
		p.monitor.Unlock()
		return ErrPoolEmpty
	}
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

func (p *limitpool) Release(ctx context.Context) {
	p.monitor.Lock()
	if p.available == p.limit {
		p.monitor.Unlock()
		panic("release exceeds limit")
	}
	p.monitor.Signal(ctx)
}
