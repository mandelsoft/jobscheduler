package uiprogress

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type Element = ppi.Element

type Ticker interface {
	Tick() bool
}

// Progress is a set of lines on a terminal
// used to display some live progress information.
// It can be used to display an arbitrary number of
// progress elements, which are independently.
// Leading elements will leave the text window
// used once they are finished.
type Progress interface {
	// UIBlocks returns the underlying
	// uiblocks.UIBlocks object used
	// to display the progress elements.
	// It can directly be used in combination
	// with progress elements.
	// But all active blocks will prohibit the
	// progress object to complete.
	UIBlocks() *uiblocks.UIBlocks

	// Done returns the done channel.
	// A Progress is done, if it is closed and
	// all progress elements are finished.
	Done() <-chan struct{}

	// Close closes the Progress. No more
	// progress elements can be added anymore.
	Close() error

	// Wait until the Progress is Done.
	// If a context.Context is given, Wait
	// also returns if the context is canceled.
	Wait(ctx context.Context) error
}

type progressImpl struct {
	lock   sync.Mutex
	blocks *uiblocks.UIBlocks
	ticker *time.Ticker

	elements []Element
}

// New creates a new Progress, which manages a terminal line range
// used to indicate progress of some actions.
// This line range is always at the end of the given
// writer, which must refer to a terminal device.
// Progress indicators are added by explicitly calling
// the appropriate constructors. They take the Progress
// they should be attached to as first argument.
func New(opt ...io.Writer) Progress {
	p := &progressImpl{
		blocks: uiblocks.New(opt...),
		ticker: time.NewTicker(time.Millisecond * 100),
	}
	go p.listen()
	return p
}

func (p *progressImpl) UIBlocks() *uiblocks.UIBlocks {
	return p.blocks
}

func (p *progressImpl) Done() <-chan struct{} {
	return p.blocks.Done()
}

func (p *progressImpl) Close() error {
	return p.blocks.Close()
}

func (p *progressImpl) Wait(ctx context.Context) error {
	return p.blocks.Wait(ctx)
}

func (p *progressImpl) listen() {
	for {
		select {
		case <-p.ticker.C:
			p.tick()
		case <-p.Done():
			return
		}
	}
}

func (p *progressImpl) tick() {
	flush := false
	for _, b := range p.blocks.Blocks() {
		if e, ok := b.Payload().(Ticker); ok {
			flush = e.Tick() || flush
		}
	}
	if flush {
		p.blocks.Flush()
	}
}
