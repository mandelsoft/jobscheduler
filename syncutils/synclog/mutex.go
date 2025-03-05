package synclog

import (
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/logging"
)

type Mutex struct {
	name string
	log  bool
	lock sync.Mutex
}

// NewMutex provides a sync.Mutex with the possibility of logging.
// Logging is enabled, if DoLog is globally set to true or
// if the optional argument for this function is set to true.
// If enabled, the ealm mandelsoft/jobscheduler/syncutils can
// be used to dynamically switch on logging with level logging.TraceLevel.
func NewMutex(name string, log ...bool) Mutex {
	return Mutex{name: name, log: general.Optional(log...)}
}

func (l *Mutex) Lock() {
	if (DoLog || l.log) && Log.Enabled(logging.TraceLevel) {
		TraceCaller("locking {{lock}}", "lock", l.name)
		l.lock.Lock()
		Log.Trace("locked", "lock", l.name)
	} else {
		l.lock.Lock()
	}
}

func (l *Mutex) LockWithReason(reason string) {
	if (DoLog || l.log) && Log.Enabled(logging.TraceLevel) {
		Log.Trace("locking {{lock}}", "lock", l.name, "reason", reason)
		l.lock.Lock()
		Log.Trace("locked", "lock", l.name)
	} else {
		l.lock.Lock()
	}
}

func (l *Mutex) Unlock() {
	if (DoLog || l.log) && Log.Enabled(logging.TraceLevel) {
		Log.Trace("unlocking {{lock}}", "lock", l.name)
	}
	l.lock.Unlock()
}

func (l *Mutex) TryLock() bool {
	b := l.lock.TryLock()
	if (DoLog || l.log) && Log.Enabled(logging.TraceLevel) {
		Log.Trace("trylock {{lock}}", "lock", l.name, "locked", b)
	}
	return b
}
