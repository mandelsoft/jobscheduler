package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

type Monitor interface {
	syncutils.Monitor
}

func NewMonitor() Monitor {
	return syncutils.NewMutexMonitor(limithandler{})
}

type limithandler struct{}

var _ utils.WaitingHandler = (*limithandler)(nil)

func (l limithandler) Release(ctx context.Context) {
	p := GetPool(ctx)
	if p != nil {
		p.Release(ctx)
	}
}

func (l limithandler) Alloc(ctx context.Context) error {
	p := GetPool(ctx)
	if p != nil {
		return p.Alloc(ctx)
	}
	return nil
}
