package syncutils

import (
	"context"
)

func Select(locker Locker, ctx context.Context) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- locker.Lock(ctx)
		close(ch)
	}()
	return ch
}
