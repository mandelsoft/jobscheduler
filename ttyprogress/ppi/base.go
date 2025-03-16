package ppi

import (
	"context"

	"github.com/mandelsoft/jobscheduler/ttyprogress/specs"
	"github.com/mandelsoft/jobscheduler/uiblocks"
)

type Container interface {
	AddBlock(b *uiblocks.UIBlock) error

	Wait(ctx context.Context) error
}

type DecoratorFunc = specs.DecoratorFunc
