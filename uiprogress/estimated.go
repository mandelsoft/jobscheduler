package uiprogress

import (
	"bytes"
	"fmt"
	"time"

	"github.com/mandelsoft/goutils/stringutils"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
	"github.com/mandelsoft/jobscheduler/uiprogress/specs"
)

// Estimated is a progress bar used to visualize the progress of an action in
// relation to a known maximum of required time
type Estimated interface {
	ppi.ProgressInterface[Estimated]

	// AppendCompleted appends the completion in percent
	// to the visualization.
	AppendCompleted(offset ...int) Estimated

	// PrependCompleted prepends the completion in percent
	// to the visualization.
	PrependCompleted(offset ...int) Estimated

	// AppendEstimated appends the estimated rest time
	AppendEstimated(offset ...int) Estimated

	// PrependEstimated prepends the estimated rest time
	PrependEstimated(offset ...int) Estimated

	// SetPending set sthe message shown before started
	SetPending(m string) Estimated

	// SetBarConfig sets a complete character configuration for the Bar.
	SetBarConfig(c BarConfig) Estimated

	// SetPredefined sets a predefined BarConfig. If the given index
	// is not defined, the operation does nothing.
	SetPredefined(i int) Estimated

	// SetFill sets the character used to indicate the
	// completed part in the visualization.
	SetFill(rune) Estimated

	// SetEmpty sets the character used to indicate the
	// pending part in the visualization.
	SetEmpty(rune) Estimated

	// SetLeftEnd sets the character used to start
	// the visualization.
	SetLeftEnd(rune) Estimated

	// SetRightEnd sets the character used to finish
	// the visualization.
	SetRightEnd(rune) Estimated

	// SetHead sets the chacter used to indicate the head of the
	// progress bar.
	SetHead(rune) Estimated

	// SetWidth sets the width of the visualization. If set
	// to zero only the prepended and appended information is shown.
	SetWidth(n uint) Estimated

	TimeEstimated() time.Duration
	TimeEstimatedString() string

	IsFinished() bool
	Set(n time.Duration) bool

	Close() error
	IsClosed() bool
}

// _Estimated represents a progress bar
type _Estimated struct {
	ppi.ProgressBase[Estimated]

	// total of the total  for the progress bar.
	total time.Duration

	// pending is the message shown before started
	pending string

	config BarConfig

	// width is the width of the progress bar.
	width uint
}

type _estimatedProtected struct {
	*_Estimated
}

func (b *_estimatedProtected) Self() Estimated {
	return b._Estimated
}

func (b *_estimatedProtected) Update() bool {
	return b._update()
}

func (b *_estimatedProtected) Visualize() (string, bool) {
	return b._visualize()
}

// NewEstimated returns a new progress bar
// based on expected execution time.
func NewEstimated(p Container, total time.Duration) Estimated {
	b := &_Estimated{
		total:   total,
		width:   Width,
		config:  BarTypes[0],
		pending: specs.Pending,
	}
	self := ppi.ProgressSelf[Estimated](&_estimatedProtected{b})
	b.ProgressBase = ppi.NewProgressBase[Estimated](self, p, 1, b.close, true)
	return b
}

// AppendCompleted appends the completion percent to the progress bar
func (b *_Estimated) AppendCompleted(offset ...int) Estimated {
	b.AppendFunc(func(b Element) string {
		return b.(*_Estimated).CompletedPercentString()
	}, offset...)
	return b
}

// PrependCompleted prepends the precent completed to the progress bar
func (b *_Estimated) PrependCompleted(offset ...int) Estimated {
	b.PrependFunc(func(b Element) string {
		return b.(*_Estimated).CompletedPercentString()
	}, offset...)
	return b
}

func (b *_Estimated) SetPending(m string) Estimated {
	b.pending = m
	return b
}

// TODO: use term width

// SetWidth sets the progress visualization width.
// The value 0 disables the visualization.
func (b *_Estimated) SetWidth(n uint) Estimated {
	b.width = n
	return b
}

func (b *_Estimated) SetBarConfig(c BarConfig) Estimated {
	b.config = c
	return b
}

func (b *_Estimated) SetPredefined(i int) Estimated {
	if c, ok := BarTypes[i]; ok {
		b.config = c
	}
	return b
}

func (b *_Estimated) SetBrackets(c Brackets) Estimated {
	b.config = b.config.SetBackets(c)
	return b
}

func (b *_Estimated) SetPredefinedBrackets(i int) Estimated {
	if c, ok := BracketTypes[i]; ok {
		b.config = b.config.SetBackets(c)
	}
	return b
}

func (b *_Estimated) SetHead(c rune) Estimated {
	b.config.Head = c
	return b
}

func (b *_Estimated) SetEmpty(c rune) Estimated {
	b.config.Empty = c
	return b
}

func (b *_Estimated) SetFill(c rune) Estimated {
	b.config.Fill = c
	return b
}

func (b *_Estimated) SetLeftEnd(c rune) Estimated {
	b.config.LeftEnd = c
	return b
}

func (b *_Estimated) SetRightEnd(c rune) Estimated {
	b.config.RightEnd = c
	return b
}

// PrependElapsed prepends the time elapsed to the begining of the bar
func (b *_Estimated) PrependEstimated(offset ...int) Estimated {
	b.PrependFunc(func(Element) string {
		return stringutils.PadLeft(b.TimeEstimatedString(), 5, ' ')
	}, offset...)
	return b
}

// AppendElapsed prepends the time elapsed to the begining of the bar
func (b *_Estimated) AppendEstimated(offset ...int) Estimated {
	b.AppendFunc(func(Element) string {
		return stringutils.PadLeft(b.TimeEstimatedString(), 5, ' ')
	}, offset...)
	return b
}

func (b *_Estimated) TimeEstimated() time.Duration {
	if b.IsStarted() {
		return b.Total() - b.TimeElapsed()
	}
	return b.Total()
}

// TimeEstimatedString returns the formatted string representation of the time elapsed
func (b *_Estimated) TimeEstimatedString() string {
	if b.IsStarted() {
		return ppi.PrettyTime(b.Total() - b.TimeElapsed())
	}
	return ""
}

// Set the current count of the bar. It returns ErrMaxCurrentReached when trying n exceeds the total value. This is atomic operation and concurrency safe.
func (b *_Estimated) Set(n time.Duration) bool {
	b.Start()

	elapsed := b.TimeElapsed()
	b.Lock.Lock()
	b.total = n
	if elapsed >= b.total {
		b.total += time.Second
	}
	b.Lock.Unlock()
	b.Flush()
	return true
}

func (b *_Estimated) close() {
	elapsed := b.TimeElapsed()

	b.Lock.Lock()
	defer b.Lock.Unlock()

	b.total = elapsed
}

func (b *_Estimated) IsFinished() bool {
	return b.IsClosed()
}

// Total returns the expected goal.
func (b *_Estimated) Total() time.Duration {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.total
}

func (b *_Estimated) _update() bool {
	return ppi.Update[Estimated](&b.ProgressBase)
}

func (b *_Estimated) _visualize() (string, bool) {
	var buf bytes.Buffer

	if !b.IsStarted() {
		return b.pending, false
	}
	// render visualization
	if b.width > 0 {
		if b.config.LeftEnd != ' ' {
			buf.Write(runeBytes(b.config.LeftEnd))
		}
		completedWidth := int(float64(b.width) * (b.CompletedPercent() / 100.00))
		// add fill and empty bits

		fill := string(b.config.Fill)
		_ = fill
		for i := 0; i < completedWidth; i++ {
			buf.Write(runeBytes(b.config.Fill))
		}
		if completedWidth > 0 {
			if completedWidth < int(b.width) {
				buf.Write(runeBytes(b.config.Head))
			}
		} else {
			buf.Write(runeBytes(b.config.Empty))
		}
		for i := 0; i < int(b.width)-completedWidth-1; i++ {
			buf.Write(runeBytes(b.config.Empty))
		}

		buf.Write(runeBytes(b.config.RightEnd))
	}
	return buf.String(), b.IsClosed()
}

// CompletedPercent return the percent completed
func (b *_Estimated) CompletedPercent() float64 {
	elapsed := b.TimeElapsed()
	total := b.total
	if total <= elapsed {
		total = elapsed
	}
	return (float64(elapsed) / float64(total)) * 100.00
}

// CompletedPercentString returns the formatted string representation of the completed percent
func (b *_Estimated) CompletedPercentString() string {
	return fmt.Sprintf("%3.f%%", b.CompletedPercent())
}

////////////////////////////////////////////////////////////////////////////////
