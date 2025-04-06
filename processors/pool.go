package processors

import (
	"context"
)

type PoolProvider interface {
	GetPool() Pool
}

type Pool interface {
	Alloc(ctx context.Context) error
	Release()
}
