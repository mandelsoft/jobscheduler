package ppi

import (
	"bytes"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/units"
)

type BaseElement interface {
	Update() bool
}

type ElemBase[T BaseInterface, I BaseElement] struct {
	Lock sync.RWMutex

	self T
	impl I

	block  *uiblocks.Block
	closer func()

	// timeStarted is time progress began.
	timeStarted time.Time
	timeElapsed time.Duration

	closed bool
}

func NewElemBase[T BaseInterface, I BaseElement](self T, impl I, b *uiblocks.UIBlocks, view int, closer func()) ElemBase[T, I] {
	if view <= 0 {
		view = 1
	}
	return ElemBase[T, I]{self: self, impl: impl, block: b.NewBlock(view).SetPayload(self), closer: closer}
}

func (b *ElemBase[T, I]) UIBlock() *uiblocks.Block {
	return b.block
}

func (b *ElemBase[T, I]) Start() {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.closed {
		return
	}

	var t time.Time
	if b.timeStarted == t {
		b.timeStarted = time.Now()
	}
}

func (b *ElemBase[T, I]) IsStarted() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	var t time.Time
	return b.timeStarted != t
}

func (b *ElemBase[T, I]) Close() error {
	err := b.close()

	if err == nil {
		b.impl.Update()
		if b.closer != nil {
			b.closer()
		}
		b.block.Close()
	}
	return err
}

func (b *ElemBase[T, I]) close() error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.closed {
		return os.ErrClosed
	}
	b.closed = true
	b.timeElapsed = time.Since(b.timeStarted)
	return nil
}

func (b *ElemBase[T, I]) IsClosed() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	return b.closed
}

// TimeElapsed returns the time elapsed
func (b *ElemBase[T, I]) TimeElapsed() time.Duration {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	if b.closed {
		return b.timeElapsed
	}
	return time.Since(b.timeStarted)
}

func PrettyTime(t time.Duration) string {
	if t == 0 {
		return ""
	}
	return units.Seconds(int(t.Truncate(time.Second) / time.Second))
}

// TimeElapsedString returns the formatted string representation of the time elapsed
func (b *ElemBase[T, I]) TimeElapsedString() string {
	return PrettyTime(b.TimeElapsed())
}

////////////////////////////////////////////////////////////////////////////////

// ProgressElement in the (protected) implementation interface.
type ProgressElement interface {
	BaseElement
	Visualize() (string, bool)
}

// ProgressBase is a base implementation for elements providing
// a line for progress information.
type ProgressBase[T ProgressInterface[T]] struct {
	ElemBase[T, ProgressElement]

	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc
}

func NewProgressBase[T ProgressInterface[T]](self T, impl ProgressElement, b *uiblocks.UIBlocks, view int, closer func()) ProgressBase[T] {
	return ProgressBase[T]{ElemBase: NewElemBase[T](self, impl, b, view, closer)}
}

func (b *ProgressBase[T]) SetFinal(m string) T {
	b.block.SetFinal(m)
	return b.self
}

// AppendFunc runs the decorator function and renders the output on the right of the progress bar
func (b *ProgressBase[T]) AppendFunc(f DecoratorFunc, offset ...int) T {
	b.Lock.Lock()
	defer b.Lock.Unlock()
	if len(offset) == 0 {
		b.appendFuncs = append(b.appendFuncs, f)
	} else {
		b.appendFuncs = slices.Insert(b.appendFuncs, offset[0], f)
	}
	return b.self
}

// PrependFunc runs decorator function and render the output left the progress bar
func (b *ProgressBase[T]) PrependFunc(f DecoratorFunc, offset ...int) T {
	b.Lock.Lock()
	defer b.Lock.Unlock()
	if len(offset) == 0 {
		b.prependFuncs = append(b.prependFuncs, f)
	} else {
		b.prependFuncs = slices.Insert(b.prependFuncs, offset[0], f)
	}
	return b.self
}

// AppendElapsed appends the time elapsed the be progress bar
func (b *ProgressBase[T]) AppendElapsed(offset ...int) T {
	b.AppendFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	return b.self
}

// PrependElapsed prepends the time elapsed to the begining of the bar
func (b *ProgressBase[T]) PrependElapsed(offset ...int) T {
	b.PrependFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	return b.self
}

func (b *ProgressBase[T]) Line() (string, bool) {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	var buf bytes.Buffer
	sep := false

	// render prepend functions to the left of the bar
	for _, f := range b.prependFuncs {
		if sep {
			buf.WriteByte(' ')
		}
		buf.Write([]byte(f(b.self)))
		sep = true
	}

	data, done := b.impl.Visualize()
	// render main function
	if len(data) > 0 {
		if sep {
			buf.WriteByte(' ')
		}
		buf.Write([]byte(data))
		sep = true
	}

	// render append functions to the right of the bar
	for _, f := range b.appendFuncs {
		if sep {
			buf.WriteByte(' ')
		}
		buf.Write([]byte(f(b.self)))
		sep = true
	}
	return buf.String(), done
}

func Update[T ProgressInterface[T]](b *ProgressBase[T]) bool {
	line, done := b.Line()

	b.block.Reset()
	b.block.Write([]byte(line + "\n"))
	if done {
		b.Close()
	}
	return true
}

func (b *ProgressBase[T]) Flush() error {
	b.impl.Update()
	return b.block.Flush()
}
