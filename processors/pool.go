package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/ctxutils"
)

type PoolProvider interface {
	GetPool() Pool
}

type Pool interface {
	PoolProvider
	Alloc(ctx context.Context) error
	Release(ctx context.Context)
}

var poolAttr = ctxutils.NewAttribute[Pool]()

func GetPool(ctx context.Context) Pool {
	return poolAttr.Get(ctx)
}

func SetPool(ctx context.Context, p Pool) context.Context {
	return poolAttr.Set(ctx, p)
}
