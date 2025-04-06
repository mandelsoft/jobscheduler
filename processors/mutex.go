package processors

import (
	"github.com/mandelsoft/jobscheduler/syncutils"
)

type Locker interface {
	syncutils.Locker
}

type Mutex interface {
	Locker
}

func NewMutex(pool PoolProvider) Mutex {
	return syncutils.NewMutex2(NewMonitor(pool))
}
