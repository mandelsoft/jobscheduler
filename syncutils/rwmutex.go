package syncutils

import (
	"context"
)

type RWMutex interface {
	Mutex

	RLock(ctx context.Context) error
	TryRLock() bool
	RUnlock()

	RLocker() Locker
}

type rwMutex struct {
	monitor Monitor
	locked  bool
	readers int
}

var _ RWMutex = (*rwMutex)(nil)

func NewRWMutex() RWMutex {
	return &rwMutex{monitor: NewMutexMonitor()}
}

func NewRWMutex2(m Monitor) RWMutex {
	return &rwMutex{monitor: m}
}

func (l *rwMutex) Lock(ctx context.Context) error {
	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked || l.readers > 0 {
		err := l.monitor.Wait(ctx)
		if err != nil {
			return err
		}
	}
	l.locked = true
	return nil
}

func (l *rwMutex) TryLock() bool {
	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked || l.readers > 0 {
		return false
	}
	l.locked = true
	return true
}

func (l *rwMutex) Unlock() {
	l.monitor.Lock()
	if !l.locked {
		l.monitor.Unlock()
		panic("unlocking unlocked rwmutex")
	}
	l.locked = false
	l.monitor.Signal(nil)
}

func (l *rwMutex) RLock(ctx context.Context) error {
	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked {
		err := l.monitor.Wait(ctx)
		if err != nil {
			return err
		}
	}
	l.readers++
	return nil
}

func (l *rwMutex) TryRLock() bool {
	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked {
		return false
	}
	l.readers++
	return true
}

func (l *rwMutex) RUnlock() {
	l.monitor.Lock()
	if l.readers == 0 {
		l.monitor.Unlock()
		panic("unlocking unlocked rwmutex")
	}
	l.readers--
	if l.readers == 0 {
		l.monitor.Signal(nil)
	} else {
		l.monitor.Unlock()
	}
}

type rlock struct {
	mutex *rwMutex
}

func (r *rlock) Lock(ctx context.Context) error {
	return r.mutex.RLock(ctx)
}

func (r *rlock) Unlock() {
	r.mutex.RUnlock()
}

func (r *rwMutex) RLocker() Locker {
	return &rlock{mutex: r}
}
