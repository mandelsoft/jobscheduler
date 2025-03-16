package ttyprogress

import (
	"bytes"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/ttyprogress/ppi"
	"github.com/mandelsoft/jobscheduler/ttyprogress/specs"
)

type BarConfig = specs.BarConfig
type Brackets = specs.Brackets

var (
	BarTypes     = specs.BarTypes
	BracketTypes = specs.BracketTypes

	// BarWidth is the default width of the progress bar
	BarWidth = specs.BarWidth

	// ErrMaxCurrentReached is error when trying to set current value that exceeds the total value
	ErrMaxCurrentReached = errors.New("errors: current value is greater total value")
)

// Bar is a progress bar used to visualize the progress of an action in
// relation to a known maximum of required work.
type Bar interface {
	specs.BarInterface

	Current() int
	IsFinished() bool
	Set(n int) bool
	Incr() bool

	Close() error
	IsClosed() bool
}

type BarDefinition struct {
	specs.BarDefinition[*BarDefinition]
}

func NewBar() *BarDefinition {
	d := &BarDefinition{}
	d.BarDefinition = specs.NewBarDefinition[*BarDefinition](specs.NewSelf[*BarDefinition](d))
	return d
}

func (d *BarDefinition) Dup() *BarDefinition {
	dup := &BarDefinition{}
	dup.BarDefinition = d.BarDefinition.Dup(dup)
	return dup
}

func (d *BarDefinition) Add(c Container) (Bar, error) {
	return newBar(c, d)
}

func (d *BarDefinition) AddWithTotal(c Container, total int) (Bar, error) {
	return newBar(c, d, total)
}

////////////////////////////////////////////////////////////////////////////////

// _Bar represents a progress bar
type _Bar struct {
	ppi.ProgressBase[Bar]

	// total of the total  for the progress bar.
	total int

	// pending is the message shown before started
	pending string

	config BarConfig

	// width is the width of the progress bar.
	width uint

	current int
}

type _barProtected struct {
	*_Bar
}

func (b *_barProtected) Self() Bar {
	return b._Bar
}

func (b *_barProtected) Update() bool {
	return b._update()
}

func (b *_barProtected) Visualize() (string, bool) {
	return b._visualize()
}

// newBar returns a new progress bar
func newBar(p Container, c specs.BarConfiguration, total ...int) (Bar, error) {
	e := &_Bar{
		total:   general.OptionalDefaulted(c.GetTotal(), total...),
		width:   c.GetWidth(),
		config:  c.GetConfig(),
		pending: c.GetPending(),
	}
	self := ppi.ProgressSelf[Bar](&_barProtected{e})
	b, err := ppi.NewProgressBase[Bar](self, p, c, 1, nil)
	if err != nil {
		return nil, err
	}
	e.ProgressBase = *b
	return e, nil
}

// Set the current count of the bar. It returns ErrMaxCurrentReached when trying n exceeds the total value. This is atomic operation and concurrency safe.
func (b *_Bar) Set(n int) bool {
	b.Start()

	b.Lock.Lock()
	if b.current >= b.total {
		b.Lock.Unlock()
		return false
	}
	if n >= b.total {
		n = b.total
	}
	b.current = n
	b.Lock.Unlock()
	b.Flush()
	return true
}

// Incr increments the current value by 1, time elapsed to current time and returns true. It returns false if the cursor has reached or exceeds total value.
func (b *_Bar) Incr() bool {
	b.Start()
	b.Lock.Lock()

	if b.incr() {
		if b.current == b.total {
			b.Lock.Unlock()
			b.Close()
		} else {
			b.Lock.Unlock()
			b.Flush()
		}
		return true
	}
	b.Lock.Unlock()
	return false
}

func (b *_Bar) IsFinished() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current == b.total
}

func (b *_Bar) incr() bool {
	if b.current == b.total {
		return false
	}

	n := b.current + 1
	b.current = n
	return true
}

// Current returns the current progress of the bar
func (b *_Bar) Current() int {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current
}

// Total returns the expected goal.
func (b *_Bar) Total() int {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.total
}

func (b *_Bar) _update() bool {
	return ppi.Update[Bar](&b.ProgressBase)
}

func runeBytes(r rune) []byte {
	return []byte(string(r))
}

func (b *_Bar) _visualize() (string, bool) {
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
	return buf.String(), b.current == b.total
}

// CompletedPercent return the percent completed
func (b *_Bar) CompletedPercent() float64 {
	return (float64(b.Current()) / float64(b.total)) * 100.00
}
