package ppi

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	// Fill is the default character representing completed progress
	Fill byte = '='

	// Head is the default character that moves when progress is updated
	Head byte = '>'

	// Empty is the default character that represents the empty progress
	Empty byte = '-'

	// LeftEnd is the default character in the left most part of the progress indicator
	LeftEnd byte = '['

	// RightEnd is the default character in the right most part of the progress indicator
	RightEnd byte = ']'

	// Width is the default width of the progress bar
	Width = uint(70)

	// ErrMaxCurrentReached is error when trying to set current value that exceeds the total value
	ErrMaxCurrentReached = errors.New("errors: current value is greater total value")
)

// Bar is a progress bar used to visualize the progress of an action in
// relation to a known maximum of required work.
type Bar interface {
	ProgressInterface[Bar]

	// AppendCompleted appends the completion in percent
	// to the visualization.
	AppendCompleted(offset ...int) Bar

	// PrependCompleted prepends the completion in percent
	// to the visualization.
	PrependCompleted(offset ...int) Bar

	// SetFill sets the character used to indicate the
	// completed part in the visualization.
	SetFill(byte) Bar

	// SetEmpty sets the character used to indicate the
	// pending part in the visualization.
	SetEmpty(byte) Bar

	// SetLeftEnd sets the character used to start
	// the visualization.
	SetLeftEnd(byte) Bar

	// SetRightEnd sets the character used to finish
	// the visualization.
	SetRightEnd(byte) Bar

	// SetHead sets the chacter used to indicate the head of the
	// progress bar.
	SetHead(byte) Bar

	// SetWidth sets the width of the visualization. If set
	// to zero only the prepended and appended information is shown.
	SetWidth(n uint) Bar

	Start()

	Current() int
	Set(n int) bool
	Incr() bool

	Close() error
	IsClosed() bool
}

// barImpl represents a progress bar
type barImpl struct {
	ProgressBase[Bar, *barImpl]

	// total of the total  for the progress bar.
	total int

	// leftEnd is character in the left most part of the progress indicator. Defaults to '['.
	leftEnd byte

	// rightEnd is character in the right most part of the progress indicator. Defaults to ']'.
	rightEnd byte

	// fill is the character representing completed progress. Defaults to '='.
	fill byte

	// head is the character that moves when progress is updated.  Defaults to '>'.
	head byte

	// empty is the character that represents the empty progress. Default is '-'.
	empty byte

	// width is the width of the progress bar.
	width uint

	current int
}

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(b Element) string

// NewBar returns a new progress bar
func NewBar(p Progress, total int) Bar {
	b := &barImpl{
		total:    total,
		width:    Width,
		leftEnd:  LeftEnd,
		rightEnd: RightEnd,
		head:     Head,
		fill:     Fill,
		empty:    Empty,
	}
	b.ProgressBase = NewProgressBase[Bar](b, p.UIBlocks(), 1, nil)
	return b
}

// AppendCompleted appends the completion percent to the progress bar
func (b *barImpl) AppendCompleted(offset ...int) Bar {
	b.AppendFunc(func(b Element) string {
		return b.(*barImpl).CompletedPercentString()
	}, offset...)
	return b
}

// PrependCompleted prepends the precent completed to the progress bar
func (b *barImpl) PrependCompleted(offset ...int) Bar {
	b.PrependFunc(func(b Element) string {
		return b.(*barImpl).CompletedPercentString()
	}, offset...)
	return b
}

// TODO: use term width

// SetWidth sets the progress visualization width.
// The value 0 disables the visualization.
func (b *barImpl) SetWidth(n uint) Bar {
	b.width = n
	return b
}

func (b *barImpl) SetHead(c byte) Bar {
	b.head = c
	return b
}

func (b *barImpl) SetEmpty(c byte) Bar {
	b.empty = c
	return b
}

func (b *barImpl) SetFill(c byte) Bar {
	b.fill = c
	return b
}

func (b *barImpl) SetLeftEnd(c byte) Bar {
	b.leftEnd = c
	return b
}

func (b *barImpl) SetRightEnd(c byte) Bar {
	b.rightEnd = c
	return b
}

// Set the current count of the bar. It returns ErrMaxCurrentReached when trying n exceeds the total value. This is atomic operation and concurrency safe.
func (b *barImpl) Set(n int) bool {
	b.Start()

	b.Lock.Lock()
	if b.current >= b.total {
		b.Lock.Unlock()
		return false
	}
	if n > b.total {
		n = b.total
	}
	b.current = n
	b.Lock.Unlock()
	b.Flush()
	return true
}

// Incr increments the current value by 1, time elapsed to current time and returns true. It returns false if the cursor has reached or exceeds total value.
func (b *barImpl) Incr() bool {
	b.Start()
	if b.incr() {
		if b.isFinished() {
			b.Close()
		} else {
			b.Flush()
		}
		return true
	}
	return false
}

func (b *barImpl) isFinished() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current == b.total
}

func (b *barImpl) incr() bool {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.current == b.total {
		return false
	}

	n := b.current + 1
	b.current = n
	return true
}

// Current returns the current progress of the bar
func (b *barImpl) Current() int {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current
}

// Total returns the expected goal.
func (b *barImpl) Total() int {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.total
}

func (b *barImpl) Visualize() (string, bool) {
	var buf bytes.Buffer

	// render visualization
	if b.width > 0 {
		if b.leftEnd != ' ' {
			buf.WriteByte(b.leftEnd)
		}
		completedWidth := int(float64(b.width) * (b.CompletedPercent() / 100.00))
		// add fill and empty bits
		for i := 0; i < completedWidth; i++ {
			buf.WriteByte(b.fill)
		}
		if completedWidth > 0 {
			if completedWidth < int(b.width) {
				buf.WriteByte(b.head)
			}
		} else {
			buf.WriteByte(b.empty)
		}
		for i := 0; i < int(b.width)-completedWidth-1; i++ {
			buf.WriteByte(b.empty)
		}

		buf.WriteByte(b.rightEnd)
	}
	return buf.String(), b.current == b.total
}

// CompletedPercent return the percent completed
func (b *barImpl) CompletedPercent() float64 {
	return (float64(b.Current()) / float64(b.total)) * 100.00
}

// CompletedPercentString returns the formatted string representation of the completed percent
func (b *barImpl) CompletedPercentString() string {
	return fmt.Sprintf("%3.f%%", b.CompletedPercent())
}

////////////////////////////////////////////////////////////////////////////////
