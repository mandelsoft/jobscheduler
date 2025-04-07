package syncutils

import (
	"github.com/mandelsoft/logging"
)

var Realm = logging.DefineRealm("mandelsoft/jobscheduler/syncutils", "mandelsoft syncutils")
var log = logging.DynamicLogger(logging.DefaultContext(), Realm)
