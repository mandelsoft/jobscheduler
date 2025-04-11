package processors

import (
	"context"
	"time"
)

func Sleep(ctx context.Context, d time.Duration) error {
	pool := GetPool(ctx)
	if pool == nil {
		if ctx == nil {
			time.Sleep(d)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(d).C:
			return nil
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	pool.Release(ctx)
	select {
	case <-ctx.Done():
		pool.Alloc(ctx)
		return ctx.Err()
	case <-time.NewTimer(d).C:
		return pool.Alloc(ctx)
	}
}
