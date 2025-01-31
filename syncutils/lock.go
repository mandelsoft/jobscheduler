package syncutils

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/logging"
)

// DoLog can be used to globally enable lock logging
// for locks provided by this package.
var DoLog = true

// DoLogCaller instructs the logging to report the caller
// of lock operations if logging is enabled.
var DoLogCaller = true

type Lock struct {
	name string
	log  bool
	lock sync.Mutex
}

// NewLock provides a sync.Lock with the possibility of logging.
// Logging is enabled, if DoLog is globally set to true or
// if the optional argument for this function is set to true.
// If enabled, the ealm mandelsoft/jobscheduler/syncutils can
// be used to dynamically switch on logging with level logging.TraceLevel.
func NewLock(name string, log ...bool) Lock {
	return Lock{name: name, log: general.Optional(log...)}
}

func (l *Lock) Lock() {
	if (DoLog || l.log) && log.Enabled(logging.TraceLevel) {
		if DoLogCaller {
			pc, file, no, ok := runtime.Caller(1)
			if ok {
				details := runtime.FuncForPC(pc)
				if details != nil {
					log.Trace("locking {{lock}}", "lock", l.name, "location", fmt.Sprintf("%s#%d", file, no), "function", details.Name())
				} else {
					log.Trace("locking {{lock}}", "lock", l.name, "location", fmt.Sprintf("%s#%d", file, no))
				}
			} else {
				log.Trace("locking {{lock}}", "lock", l.name)
			}
		} else {
			log.Trace("locking {{lock}}", "lock", l.name)
		}
		l.lock.Lock()
		log.Trace("locked", "lock", l.name)
	} else {
		l.lock.Lock()
	}
}

func (l *Lock) LockWithReason(reason string) {
	log.Trace("locking {{lock}}", "lock", l.name, "reason", reason)
	l.lock.Lock()
	log.Trace("locked", "lock", l.name)
}

func (l *Lock) Unlock() {
	log.Trace("unlocking {{lock}}", "lock", l.name)
	l.lock.Unlock()
}

func (l *Lock) TryLock() bool {
	b := l.lock.TryLock()
	log.Trace("trylock {{lock}}", "lock", l.name, "locked", b)
	return b
}
