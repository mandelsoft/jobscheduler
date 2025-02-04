package uiprogress

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

// Bar represents a progress bar
type Bar struct {
	ProgressBase[*Bar]

	// Total of the total  for the progress bar.
	Total int

	// LeftEnd is character in the left most part of the progress indicator. Defaults to '['.
	LeftEnd byte

	// RightEnd is character in the right most part of the progress indicator. Defaults to ']'.
	RightEnd byte

	// Fill is the character representing completed progress. Defaults to '='.
	Fill byte

	// Head is the character that moves when progress is updated.  Defaults to '>'.
	Head byte

	// Empty is the character that represents the empty progress. Default is '-'.
	Empty byte

	// Width is the width of the progress bar.
	Width uint

	current int
}

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(b Element) string

// NewBar returns a new progress bar
func NewBar(p *Progress, total int) *Bar {
	b := &Bar{
		Total:    total,
		Width:    Width,
		LeftEnd:  LeftEnd,
		RightEnd: RightEnd,
		Head:     Head,
		Fill:     Fill,
		Empty:    Empty,
	}
	b.ProgressBase = NewProgressBase(b, p.blocks, ProgressBaseOptions{})
	return b
}

// AppendCompleted appends the completion percent to the progress bar
func (b *Bar) AppendCompleted(offset ...int) *Bar {
	b.AppendFunc(func(b Element) string {
		return b.(*Bar).CompletedPercentString()
	}, offset...)
	return b
}

// PrependCompleted prepends the precent completed to the progress bar
func (b *Bar) PrependCompleted(offset ...int) *Bar {
	b.PrependFunc(func(b Element) string {
		return b.(*Bar).CompletedPercentString()
	}, offset...)
	return b
}

// TODO: use term width

// SetWidth sets the progress visualization width.
// The value 0 disables the visualization.
func (b *Bar) SetWidth(n uint) *Bar {
	b.Width = n
	return b
}

// Set the current count of the bar. It returns ErrMaxCurrentReached when trying n exceeds the total value. This is atomic operation and concurrency safe.
func (b *Bar) Set(n int) bool {
	b.Start()

	b.Lock.Lock()
	if b.current >= b.Total {
		b.Lock.Unlock()
		return false
	}
	if n > b.Total {
		n = b.Total
	}
	b.current = n
	b.Lock.Unlock()
	b.Flush()
	return true
}

// Incr increments the current value by 1, time elapsed to current time and returns true. It returns false if the cursor has reached or exceeds total value.
func (b *Bar) Incr() bool {
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

func (b *Bar) isFinished() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current == b.Total
}

func (b *Bar) incr() bool {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if b.current == b.Total {
		return false
	}

	n := b.current + 1
	b.current = n
	return true
}

// Current returns the current progress of the bar
func (b *Bar) Current() int {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current
}

func (b *Bar) Visualize() (string, bool) {
	var buf bytes.Buffer

	// render visualization
	if b.Width > 0 {
		if b.LeftEnd != ' ' {
			buf.WriteByte(b.LeftEnd)
		}
		completedWidth := int(float64(b.Width) * (b.CompletedPercent() / 100.00))
		// add fill and empty bits
		for i := 0; i < completedWidth; i++ {
			buf.WriteByte(b.Fill)
		}
		if completedWidth > 0 {
			if completedWidth < int(b.Width) {
				buf.WriteByte(b.Head)
			}
		} else {
			buf.WriteByte(b.Empty)
		}
		for i := 0; i < int(b.Width)-completedWidth-1; i++ {
			buf.WriteByte(b.Empty)
		}

		buf.WriteByte(b.RightEnd)
	}
	return buf.String(), b.current == b.Total
}

// CompletedPercent return the percent completed
func (b *Bar) CompletedPercent() float64 {
	return (float64(b.Current()) / float64(b.Total)) * 100.00
}

// CompletedPercentString returns the formatted string representation of the completed percent
func (b *Bar) CompletedPercentString() string {
	return fmt.Sprintf("%3.f%%", b.CompletedPercent())
}

////////////////////////////////////////////////////////////////////////////////
