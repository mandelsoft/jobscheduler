package processors

import (
	"context"
)

func ReceiveFromChannel[T any](ctx context.Context, c <-chan T) (T, bool, error) {
	var _nil T

	pool := GetPool(ctx)
	if pool == nil {
		v, ok := <-c
		return v, ok, nil
	}

	select {
	case v, ok := <-c:
		return v, ok, nil
	default:
	}
	pool.Release(ctx)
	select {
	case v, ok := <-c:
		err := pool.Alloc(ctx)
		return v, ok, err
	case <-ctx.Done():
		pool.Alloc(ctx)
		return _nil, false, ctx.Err()
	}
}

func SendToChannel[T any](ctx context.Context, c chan<- T, v T) error {
	pool := GetPool(ctx)
	if pool == nil {
		c <- v
		return nil
	}

	select {
	case c <- v:
		return nil
	default:
	}
	pool.Release(ctx)
	select {
	case c <- v:
		return pool.Alloc(ctx)
	case <-ctx.Done():
		pool.Alloc(ctx)
		return ctx.Err()
	}
}
