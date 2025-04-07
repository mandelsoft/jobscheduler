package processors

import (
	"context"
	"time"
)

func NotFunc(f func() bool) func() bool {
	return func() bool { return !f() }
}

func Sleep(p Pool, d time.Duration, ctx context.Context) error {
	p.Release(nil)
	if ctx == nil {
		time.Sleep(d)
	} else {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(d).C:
		}
	}
	return p.Alloc(ctx)
}
