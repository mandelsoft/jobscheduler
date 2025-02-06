package uiprogress

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mandelsoft/jobscheduler/uiblocks"
)

type Group interface {
	Container

	SetFollowUpGap(gap string) Group
	Gap() string

	ProgressInterface[Group]
	SetSpeed(int) Group
}

type _group struct {
	lock     sync.RWMutex
	parent   Container
	gap      string
	pgap     string
	followup string

	spinner Spinner
	blocks  []*uiblocks.UIBlock
	closed  bool
}

var _ Group = (*_group)(nil)

func NewGroup(p Container, gap string, set int) Group {
	g := &_group{
		parent:   p,
		gap:      gap,
		followup: gap,
	}
	if pg, ok := p.(Group); ok {
		g.pgap = pg.Gap()
	}

	g.spinner = NewSpinner(g, set)
	if g.spinner == nil {
		return nil
	}
	return g
}

func (g *_group) SetFollowUpGap(gap string) Group {
	g.followup = gap
	return g
}

func (g *_group) NewBlock(view ...int) *uiblocks.UIBlock {
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
		b = g.blocks[0].UIBlocks().AppendBlock(n, view...).
			SetGap(g.pgap + g.gap).SetFollowUpGap(g.pgap + g.followup)
	}
	if b != nil {
		g.blocks = append(g.blocks, b)
		g.blocks[0].SetNext(b)
	}
	return b
}

func (g *_group) Gap() string {
	return g.pgap + g.followup
}

func (g *_group) TimeElapsed() time.Duration {
	return g.spinner.TimeElapsed()
}

func (g *_group) TimeElapsedString() string {
	return g.spinner.TimeElapsedString()
}

func (g *_group) Start() {
	g.spinner.Start()
}

func (g *_group) IsStarted() bool {
	return g.spinner.IsStarted()
}

func (g *_group) Close() error {
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
		g.spinner.Close()
	}()
	return nil
}

func (g *_group) IsClosed() bool {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.closed
}

func (g *_group) Wait(ctx context.Context) error {
	return g.blocks[0].Wait(ctx)
}

////////////////////////////////////////////////////////////////////////////////

func (g *_group) SetFinal(m string) Group {
	g.spinner.SetFinal(m)
	return g
}

func (g *_group) SetColor(col *color.Color) Group {
	g.spinner.SetColor(col)
	return g
}

func (g *_group) AppendFunc(f DecoratorFunc, offset ...int) Group {
	g.spinner.AppendFunc(f, offset...)
	return g
}

func (g *_group) PrependFunc(f DecoratorFunc, offset ...int) Group {
	g.spinner.PrependFunc(f, offset...)
	return g
}

func (g *_group) AppendElapsed(offset ...int) Group {
	g.spinner.AppendElapsed(offset...)
	return g
}

func (g *_group) PrependElapsed(offset ...int) Group {
	g.spinner.PrependElapsed(offset...)
	return g
}

func (g *_group) SetSpeed(i int) Group {
	g.spinner.SetSpeed(i)
	return g
}
