package ppi

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mandelsoft/jobscheduler/uiblocks"
)

type Gapped interface {
	Gap() string
}

type GroupBase[I any, T ProgressInterface[T]] struct {
	self     I
	lock     sync.RWMutex
	parent   Container
	gap      string
	pgap     string
	followup string

	main   T
	blocks []*uiblocks.UIBlock
	closed bool
}

func NewGroupBase[I any, T ProgressInterface[T]](p Container, self I, gap string, main func(base *GroupBase[I, T]) (T, bool)) (*GroupBase[I, T], T) {
	g := &GroupBase[I, T]{
		self:     self,
		parent:   p,
		gap:      gap,
		followup: gap,
	}
	if pg, ok := p.(Gapped); ok {
		g.pgap = pg.Gap()
	}

	if m, ok := main(g); !ok {
		return nil, g.main
	} else {
		g.main = m
		return g, g.main
	}
}

func (g *GroupBase[I, T]) SetFollowUpGap(gap string) I {
	g.followup = gap
	return g.self
}

func (g *GroupBase[I, T]) NewBlock(view ...int) *uiblocks.UIBlock {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.closed {
		return nil
	}

	var b *uiblocks.UIBlock
	if len(g.blocks) == 0 {
		b = g.parent.NewBlock(view...)
		b.SetGap(g.pgap)
	} else {
		g.Start()
		n := g.blocks[0]
		for n.Next() != nil && n.Next() != n {
			n = n.Next()
		}
		b = g.blocks[0].UIBlocks().NewAppendedBlock(n, view...).
			SetGap(g.pgap + g.gap).SetFollowUpGap(g.pgap + g.followup)
	}
	if b != nil {
		g.blocks = append(g.blocks, b)
		g.blocks[0].SetNext(b)
	}
	return b
}

func (g *GroupBase[I, T]) Flush() error {
	return g.main.Flush()
}

func (g *GroupBase[I, T]) Gap() string {
	return g.pgap + g.followup
}

func (g *GroupBase[I, T]) TimeElapsed() time.Duration {
	return g.main.TimeElapsed()
}

func (g *GroupBase[I, T]) TimeElapsedString() string {
	return g.main.TimeElapsedString()
}

func (g *GroupBase[I, T]) Start() Element {
	return g.main.Start()
}

func (g *GroupBase[I, T]) IsStarted() bool {
	return g.main.IsStarted()
}

func (g *GroupBase[I, T]) Close() error {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.closed {
		return os.ErrClosed
	}
	g.closed = true

	go func() {
		for _, b := range g.blocks[1:] {
			b.Wait(nil)
		}
		g.main.Close()
	}()
	return nil
}

func (g *GroupBase[I, T]) IsClosed() bool {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.closed
}

func (g *GroupBase[I, T]) Wait(ctx context.Context) error {
	return g.blocks[0].Wait(ctx)
}

////////////////////////////////////////////////////////////////////////////////

func (g *GroupBase[I, T]) SetFinal(m string) I {
	g.main.SetFinal(m)
	return g.self
}

func (g *GroupBase[I, T]) SetColor(col *color.Color) I {
	g.main.SetColor(col)
	return g.self
}

func (g *GroupBase[I, T]) AppendFunc(f DecoratorFunc, offset ...int) I {
	g.main.AppendFunc(f, offset...)
	return g.self
}

func (g *GroupBase[I, T]) PrependFunc(f DecoratorFunc, offset ...int) I {
	g.main.PrependFunc(f, offset...)
	return g.self
}

func (g *GroupBase[I, T]) AppendElapsed(offset ...int) I {
	g.main.AppendElapsed(offset...)
	return g.self
}

func (g *GroupBase[I, T]) PrependElapsed(offset ...int) I {
	g.main.PrependElapsed(offset...)
	return g.self
}
