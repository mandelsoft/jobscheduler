package scheduler

import (
	"github.com/mandelsoft/logging"
)

var Realm = logging.DefineRealm("mandelsoft/jobscheduler/scheduler", "mandelsoft job scheduler")
var log = logging.DynamicLogger(logging.DefaultContext(), Realm)
