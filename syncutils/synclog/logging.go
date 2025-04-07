package synclog

import (
	"fmt"
	"runtime"

	"github.com/mandelsoft/jobscheduler/syncutils/utils"
	"github.com/mandelsoft/logging"
)

var Realm = utils.Realm
var Log = logging.DynamicLogger(logging.DefaultContext(), Realm)

// DoLog can be used to globally enable lock logging
// for locks provided by this package.
var DoLog = true

// DoLogCaller instructs the logging to report the caller
// of lock operations if logging is enabled.
var DoLogCaller = true

func TraceCaller(msg string, args ...interface{}) {
	if DoLogCaller {
		pc, file, no, ok := runtime.Caller(2)
		if ok {
			details := runtime.FuncForPC(pc)
			if details != nil {
				Log.Trace(msg, append([]any{"location", fmt.Sprintf("%s#%d", file, no), "function", details.Name()}, args...)...)
			} else {
				Log.Trace(msg, append([]any{"location", fmt.Sprintf("%s#%d", file, no)}, args...)...)
			}
		} else {
			Log.Trace(msg, args...)
		}
	} else {
		Log.Trace(msg, args...)
	}
}
