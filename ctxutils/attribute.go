package ctxutils

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
)

type Attribute[T any] interface {
	Get(context.Context) T
	Set(context.Context, T) context.Context
}

type attribute[T any] struct {
	key *string
}

func NewAttribute[T any]() Attribute[T] {
	return &attribute[T]{key: generics.Pointer(generics.TypeOf[T]().String())}

}

func (a *attribute[T]) Get(ctx context.Context) T {
	var _nil T
	if ctx == nil {
		return _nil
	}
	return generics.Cast[T](ctx.Value(a.key))
}

func (a *attribute[T]) Set(ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, a.key, v)
}
