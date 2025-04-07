package syncutils

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
	"github.com/mandelsoft/logging"
)

type Locker interface {
	Lock(ctx context.Context) error
	Unlock()
}

type Mutex interface {
	Locker
	TryLock() bool
}

type mutex struct {
	monitor Monitor
	log     bool
	name    string
	locked  bool
}

func options(args ...any) (string, bool, utils.WaitingHandler) {
	log := false
	name := ""
	var handler utils.WaitingHandler
	for _, e := range args {
		switch d := e.(type) {
		case bool:
			log = d
		case string:
			name = d
		case utils.WaitingHandler:
			handler = d
		}
	}
	return name, log, handler
}

// NewMutex create a new Mutex, optional
// arguments are of type bool for enable logging
// string for the name of the Mutex, or utils.WaitingHandler
// for the  created monitor.
func NewMutex(args ...any) Mutex {
	name, log, h := options(args...)
	return &mutex{monitor: NewMutexMonitor(h), log: log, name: name}
}

func NewMutex2(m Monitor, args ...any) Mutex {
	name, log, _ := options(args...)
	return &mutex{monitor: m, log: log, name: name}
}

func (l *mutex) Lock(ctx context.Context) error {
	if (synclog.DoLog || l.log) && synclog.Log.Enabled(logging.TraceLevel) {
		synclog.TraceCaller("locking {{lock}}", "lock", l.name)
	}

	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked {
		err := l.monitor.Wait(ctx)
		if err != nil {
			synclog.Log.Trace("locking {{lock}} failed", "lock", l.name, "error", err)
			return err
		}
	}
	l.locked = true
	if (synclog.DoLog || l.log) && synclog.Log.Enabled(logging.TraceLevel) {
		synclog.Log.Trace("locking {{lock}} locked", "lock", l.name)
	}
	return nil
}

func (l *mutex) TryLock() bool {
	l.monitor.Lock()
	defer l.monitor.Unlock()

	if l.locked {
		if (synclog.DoLog || l.log) && synclog.Log.Enabled(logging.TraceLevel) {
			synclog.Log.Trace("trylock {{lock}} failed", "lock", l.name)
		}
		return false
	}
	l.locked = true
	if (synclog.DoLog || l.log) && synclog.Log.Enabled(logging.TraceLevel) {
		synclog.Log.Trace("trylock {{lock}} locked", "lock", l.name)
	}
	return true
}

func (l *mutex) Unlock() {
	if (synclog.DoLog || l.log) && synclog.Log.Enabled(logging.TraceLevel) {
		synclog.Log.Trace("unlocking {{lock}}", "lock", l.name)
	}
	l.monitor.Lock()
	if !l.locked {
		l.monitor.Unlock()
		panic("unlocking unlocked mutex")
	}
	l.locked = false
	l.monitor.Signal(nil)
}
