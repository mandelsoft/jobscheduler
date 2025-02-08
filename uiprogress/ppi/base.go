package ppi

import (
	"bytes"
	"context"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/units"
)

type Container interface {
	NewBlock(view ...int) *uiblocks.UIBlock
	Wait(ctx context.Context) error
}

type BaseProtected[I any] interface {
	Protected[I]
	Update() bool
}

func BaseSelf[I BaseInterface, P BaseProtected[I]](p P) Self[I, P] {
	return Self[I, P]{p}
}

type ElemBase[I BaseInterface, P BaseProtected[I]] struct {
	Lock sync.RWMutex

	self Self[I, P]

	block  *uiblocks.UIBlock
	closer func()

	// timeStarted is time progress began.
	timeStarted time.Time
	timeElapsed time.Duration

	closed bool
}

func NewElemBase[I BaseInterface, P BaseProtected[I]](self Self[I, P], p Container, view int, closer func()) ElemBase[I, P] {
	if view <= 0 {
		view = 1
	}
	return ElemBase[I, P]{self: self, block: p.NewBlock(view).SetPayload(self.Self()), closer: closer}
}

func (b *ElemBase[I, P]) SetFinal(m string) I {
	b.block.SetFinal(m)
	return b.self.Self()
}

func (b *ElemBase[I, P]) UIBlock() *uiblocks.UIBlock {
	return b.block
}

func (b *ElemBase[I, P]) Start() BaseInterface {
	if b.start() {
		b.self.Protected().Update()
		b.block.Flush()
		return b.self.Self()
	}
	return nil
}

func (b *ElemBase[I, P]) start() bool {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	var t time.Time
	if b.closed || b.timeStarted != t {
		return false
	}
	b.timeStarted = time.Now()
	return true
}

func (b *ElemBase[I, P]) IsStarted() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	var t time.Time
	return b.timeStarted != t
}

func (b *ElemBase[I, P]) Close() error {
	err := b.close()

	if err == nil {
		b.self.Protected().Update()
		if b.closer != nil {
			b.closer()
		}
		b.block.Close()
	}
	return err
}

func (b *ElemBase[I, P]) close() error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.closed {
		return os.ErrClosed
	}
	b.closed = true
	b.timeElapsed = time.Since(b.timeStarted)
	return nil
}

func (b *ElemBase[I, P]) IsClosed() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	return b.closed
}

func (b *ElemBase[I, P]) Wait(ctx context.Context) error {
	return b.block.Wait(ctx)
}

// TimeElapsed returns the time elapsed
func (b *ElemBase[I, P]) TimeElapsed() time.Duration {
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
func (b *ElemBase[I, P]) TimeElapsedString() string {
	return PrettyTime(b.TimeElapsed())
}

////////////////////////////////////////////////////////////////////////////////

// ProgressProtected in the (protected) implementation interface for progress
// indicators.
type ProgressProtected[I any] interface {
	BaseProtected[I]
	Visualize() (string, bool)
}

func ProgressSelf[I ProgressInterface[I]](impl ProgressProtected[I]) Self[I, ProgressProtected[I]] {
	return Self[I, ProgressProtected[I]]{impl}
}

// ProgressBase is a base implementation for elements providing
// a line for progress information.
type ProgressBase[T ProgressInterface[T]] struct {
	ElemBase[T, ProgressProtected[T]]

	color *color.Color

	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc
	tick         bool
}

func NewProgressBase[T ProgressInterface[T]](self Self[T, ProgressProtected[T]], p Container, view int, closer func()) ProgressBase[T] {
	return ProgressBase[T]{ElemBase: NewElemBase[T, ProgressProtected[T]](self, p, view, closer)}
}

func (b *ProgressBase[T]) Tick() bool {
	if b.tick && !b.closed {
		b.self.Protected().Update()
		return true
	}
	return false
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
	return b.self.Self()
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
	return b.self.Self()
}

// SetColor appends the time elapsed the be progress bar
func (b *ProgressBase[T]) SetColor(col *color.Color) T {
	b.color = col
	return b.self.Self()
}

// AppendElapsed appends the time elapsed the be progress bar
func (b *ProgressBase[T]) AppendElapsed(offset ...int) T {
	b.AppendFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	b.tick = true
	return b.self.Self()
}

// PrependElapsed prepends the time elapsed to the begining of the bar
func (b *ProgressBase[T]) PrependElapsed(offset ...int) T {
	b.PrependFunc(func(Element) string {
		return strutils.PadLeft(b.TimeElapsedString(), 5, ' ')
	}, offset...)
	b.tick = true
	return b.self.Self()
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
		buf.Write([]byte(f(b.self.Self())))
		sep = true
	}

	data, done := b.self.Protected().Visualize()
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
		buf.Write([]byte(f(b.self.Self())))
		sep = true
	}

	if b.color != nil {
		return b.color.Sprint(buf.String()), done
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
	b.self.Protected().Update()
	return b.block.Flush()
}
