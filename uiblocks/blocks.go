package uiblocks

import (
	"context"
	"io"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/mandelsoft/goutils/general"
)

// UIBlocks is a buffered the writer that updates the terminal. The contents of writer will be flushed on a timed interval or when Flush is called.
type UIBlocks struct {
	lock sync.Mutex

	// out is the writer to write to
	out       io.Writer
	termWidth int

	overFlowHandled bool

	blocks    []*Block
	lineCount int

	closed bool
	done   chan struct{}

	timer   *time.Timer
	pending bool
}

// New returns a new UIBlocks with defaults
func New(opt ...io.Writer) *UIBlocks {
	w := &UIBlocks{
		out:   general.OptionalDefaulted[io.Writer](os.Stdout, opt...),
		done:  make(chan struct{}),
		timer: time.NewTimer(0),
	}

	termWidth, _ := getTermSize()
	if termWidth != 0 {
		w.termWidth = termWidth
		w.overFlowHandled = true
	}
	go w.listen()
	return w
}

func (w *UIBlocks) requestFlush() {
	if w.pending {
		return
	}
	w.pending = true
	w.timer.Reset(time.Millisecond * 250)
}

func (w *UIBlocks) Done() <-chan struct{} {
	return w.done
}

func (w *UIBlocks) listen() {
	for {
		select {
		case <-w.done:
			return
		case <-w.timer.C:
			w.Flush()
		}
	}
}

func (w *UIBlocks) NewBlock(view ...int) *Block {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.closed {
		return nil
	}
	b := &Block{blocks: w, view: general.OptionalDefaulted(DefaultView, view...)}
	w.blocks = append(w.blocks, b)
	return b
}

func (w *UIBlocks) Blocks() []*Block {
	w.lock.Lock()
	defer w.lock.Unlock()
	return slices.Clone(w.blocks)
}

func (w *UIBlocks) TermWidth() int {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.termWidth
}

func (w *UIBlocks) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.closed {
		return os.ErrClosed
	}
	w.closed = true
	return nil
}

func (w *UIBlocks) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return nil
	}
}

func (w *UIBlocks) discardBlock() error {
	discarded := false
	for len(w.blocks) > 0 && w.blocks[0].closed {
		if !discarded {
			w.clearLines()
			discarded = true
		}
		w.blocks[0].flush(true)
		w.blocks = w.blocks[1:]
	}
	if discarded {
		err := w.flush()
		if w.closed && len(w.blocks) == 0 {
			close(w.done)
		}
		return err
	}
	return nil
}

func (w *UIBlocks) Flush() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.clearLines()
	w.timer.Stop()
	w.pending = false
	return w.flush()
}

func (w *UIBlocks) flush() error {
	lines := 0
	for _, b := range w.blocks {
		l, err := b.flush(false)
		lines += l
		if err != nil {
			return err
		}
	}
	w.lineCount = lines
	return err
}
