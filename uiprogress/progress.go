package uiprogress

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/mandelsoft/jobscheduler/uiblocks"
)

type Element interface {
}

type Ticker interface {
	Tick() bool
}

type Progress struct {
	lock   sync.Mutex
	blocks *uiblocks.UIBlocks
	ticker *time.Ticker

	elements []Element
}

func New(opt ...io.Writer) *Progress {
	p := &Progress{
		blocks: uiblocks.New(opt...),
		ticker: time.NewTicker(time.Millisecond * 100),
	}
	go p.listen()
	return p
}

func (p *Progress) Done() <-chan struct{} {
	return p.blocks.Done()
}

func (p *Progress) Close() error {
	return p.blocks.Close()
}

func (p *Progress) Wait(ctx context.Context) error {
	return p.blocks.Wait(ctx)
}

func (p *Progress) listen() {
	for {
		select {
		case <-p.ticker.C:
			p.tick()
		case <-p.Done():
			return
		}
	}
}

func (p *Progress) tick() {
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
