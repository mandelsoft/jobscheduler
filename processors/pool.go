package processors

import (
	"context"
)

type Pool interface {
	Alloc(ctx context.Context) error
	Release()
}
