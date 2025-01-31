package queue

import (
	"github.com/mandelsoft/logging"
)

var Realm = logging.DefineRealm("mandelsoft/jobscheduler/queue", "request queue")
var log = logging.DynamicLogger(logging.DefaultContext(), Realm)
