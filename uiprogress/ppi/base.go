package ppi

import (
	"bytes"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/units"
)

type BaseElement interface {
	Update() bool
}

type ElemBase[T BaseElement] struct {
	Lock sync.RWMutex

	self   T
	block  *uiblocks.Block
	closer func()

	// timeStarted is time progress began.
	timeStarted time.Time
	timeElapsed time.Duration

	closed bool
}

func NewElemBase[T BaseElement](self T, b *uiblocks.UIBlocks, view int, closer func()) ElemBase[T] {
	if view <= 0 {
		view = 1
	}
	return ElemBase[T]{self: self, block: b.NewBlock(view).SetPayload(self), closer: closer}
}

func (b *ElemBase[T]) UIBlock() *uiblocks.Block {
	return b.block
}

func (b *ElemBase[T]) Start() {
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

func (b *ElemBase[T]) IsStarted() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	var t time.Time
	return b.timeStarted != t
}

func (b *ElemBase[T]) Close() error {
	err := b.close()

	if err == nil {
		b.self.Update()
		if b.closer != nil {
			b.closer()
		}
		b.block.Close()
	}
	return err
}

func (b *ElemBase[T]) close() error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.closed {
		return os.ErrClosed
	}
	b.closed = true
	b.timeElapsed = time.Since(b.timeStarted)
	return nil
}

func (b *ElemBase[T]) IsClosed() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	return b.closed
}

// TimeElapsed returns the time elapsed
func (b *ElemBase[T]) TimeElapsed() time.Duration {
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
func (b *ElemBase[T]) TimeElapsedString() string {
	return PrettyTime(b.TimeElapsed())
}

////////////////////////////////////////////////////////////////////////////////

// ProgressElement in the (protected) implementation interface.
type ProgressElement interface {
	BaseElement
	Start()
	IsClosed() bool
	Visualize() (string, bool)
}

// ProgressImplementation in the complete implementation interface.
type ProgressImplementation[P ProgressInterface[P]] interface {
	ProgressInterface[P]
	ProgressElement
}

// ProgressBase is a base implementation for elements providing
// a line for progress information.
type ProgressBase[P ProgressInterface[P], T ProgressImplementation[P]] struct {
	ElemBase[T]

	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc
}

func NewProgressBase[P ProgressInterface[P], T ProgressImplementation[P]](self T, b *uiblocks.UIBlocks, view int, closer func()) ProgressBase[P, T] {
	return ProgressBase[P, T]{ElemBase: NewElemBase[T](self, b, view, closer)}
}

func (b *ProgressBase[P, T]) SetFinal(m string) P {
	b.block.SetFinal(m)
	return generics.Cast[P](b.self)
}

// AppendFunc runs the decorator function and renders the output on the right of the progress bar
func (b *ProgressBase[P, T]) AppendFunc(f DecoratorFunc, offset ...int) P {
	b.Lock.Lock()
	defer b.Lock.Unlock()
	if len(offset) == 0 {
		b.appendFuncs = append(b.appendFuncs, f)
	} else {
		b.appendFuncs = slices.Insert(b.appendFuncs, offset[0], f)
	}
	return generics.Cast[P](b.self)
}

// PrependFunc runs decorator function and render the output left the progress bar
func (b *ProgressBase[P, T]) PrependFunc(f DecoratorFunc, offset ...int) P {
	b.Lock.Lock()
	defer b.Lock.Unlock()
	if len(offset) == 0 {
		b.prependFuncs = append(b.prependFuncs, f)
	} else {
		b.prependFuncs = slices.Insert(b.prependFuncs, offset[0], f)
	}
	return generics.Cast[P](b.self)
}

// AppendElapsed appends the time elapsed the be progress bar
func (b *ProgressBase[P, T]) AppendElapsed(offset ...int) P {
	b.AppendFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	return generics.Cast[P](b.self)
}

// PrependElapsed prepends the time elapsed to the begining of the bar
func (b *ProgressBase[P, T]) PrependElapsed(offset ...int) P {
	b.PrependFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	return generics.Cast[P](b.self)
}

func (b *ProgressBase[P, T]) Line() (string, bool) {
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

	data, done := b.self.Visualize()
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

func (b *ProgressBase[P, T]) Update() bool {
	line, done := b.Line()

	b.block.Reset()
	b.block.Write([]byte(line + "\n"))
	if done {
		b.Close()
	}
	return true
}

func (b *ProgressBase[P, T]) Flush() error {
	b.self.Update()
	return b.block.Flush()
}
