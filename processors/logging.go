package processors

import (
	"github.com/mandelsoft/logging"
)

var Realm = logging.DefineRealm("mandelsoft/jobscheduler/processors", "mandelsoft processor management")
var log = logging.DynamicLogger(logging.DefaultContext(), Realm)
