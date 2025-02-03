package uiprogress

import (
	"bytes"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/units"
)

type ElemBase[T Element] struct {
	Lock sync.RWMutex

	self  T
	block *uiblocks.Block
}

func NewElemBase[T any](self T, b *uiblocks.Block) ElemBase[T] {
	return ElemBase[T]{self: self, block: b}
}

func (b ElemBase[T]) UIBlock() *uiblocks.Block {
	return b.block
}

////////////////////////////////////////////////////////////////////////////////

type MainFunc func(Element) (string, bool)

type ProgressElement interface {
	Element
	Start()
	Flush() error
	Visualize() (string, bool)
}

type ProgressBase[T ProgressElement] struct {
	ElemBase[T]

	appendFuncs  []DecoratorFunc
	prependFuncs []DecoratorFunc

	// timeStarted is time progress began.
	timeStarted time.Time
	finished    bool
}

type TimeElapsed interface {
	TimeElapsed() time.Duration
}

func NewProgressBase[T ProgressElement](self T, b *uiblocks.UIBlocks, view ...int) ProgressBase[T] {
	return ProgressBase[T]{ElemBase: NewElemBase[T](self, b.NewBlock(general.OptionalDefaulted(1, view...)).SetPayload(self))}
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

func (b *ProgressBase[T]) Start() {
	b.Lock.Lock()
	b.start()
	b.Lock.Unlock()
	b.self.Flush()
}

func (b *ProgressBase[T]) start() {
	var t time.Time

	if b.timeStarted == t {
		b.timeStarted = time.Now()
	}
}

// TimeElapsed returns the time elapsed
func (b *ProgressBase[T]) TimeElapsed() time.Duration {
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	return time.Since(b.timeStarted)
}

func PrettyTime(t time.Duration) string {
	if t == 0 {
		return ""
	}
	return units.Seconds(int(t.Truncate(time.Second) / time.Second))
}

// TimeElapsedString returns the formatted string represenation of the time elapsed
func (b *ProgressBase[T]) TimeElapsedString() string {
	return PrettyTime(b.TimeElapsed())
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

func (b *ProgressBase[T]) Flush() error {
	line, done := b.Line()

	b.block.Reset()
	_, err := b.block.Write([]byte(line + "\n"))

	err = b.block.Flush()
	if done {
		b.block.Close()
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////

func Message(m string) DecoratorFunc {
	return func(element Element) string {
		return m
	}
}

func Amount(unit ...units.Unit) func(Element) string {
	u := general.OptionalDefaulted(units.Plain, unit...)
	return func(e Element) string {
		return fmt.Sprintf("(%s/%s)", u(e.(*Bar).Current()), u(e.(*Bar).Total))
	}
}
