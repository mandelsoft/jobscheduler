package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils"
)

type Monitor interface {
	syncutils.Monitor
}

func NewMonitor(pool PoolProvider) Monitor {
	return &monitor{syncutils.NewMutexMonitor(), pool.GetPool()}
}

type monitor struct {
	monitor syncutils.Monitor
	pool    Pool
}

func (m *monitor) Lock() {
	m.monitor.Lock()
}

func (m *monitor) Unlock() {
	m.monitor.Unlock()
}

func (m *monitor) Wait(ctx context.Context) error {
	m.pool.Release()
	// if wait is regularily signaled the processor and the lock
	// is transferred
	return m.monitor.Wait(ctx)
}

// Signal continues a waiting routine.
// The lock and the processor are transferred.
// The actual routine reallocates a processor.
// If there is none to signal/deblock, the lock
// is released and the processor is kept.
func (m *monitor) Signal() bool {
	ok := m.monitor.Signal()
	if ok {
		m.pool.Alloc(nil) // Todo: handle abort
	}
	return ok
}

func (m *monitor) SignalAll() bool {
	return m.monitor.SignalAll()
}

func (m *monitor) HasWaiting() bool {
	return m.monitor.HasWaiting()
}
