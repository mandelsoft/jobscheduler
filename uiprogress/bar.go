package uiprogress

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type BarConfig struct {
	// Fill is the default character representing completed progress
	Fill rune
	// Head is the default character that moves when progress is updated
	Head rune
	// Empty is the default character that represents the empty progress
	Empty rune
	// LeftEnd is the default character in the left most part of the progress indicator
	LeftEnd rune
	// RightEnd is the default character in the right most part of the progress indicator
	RightEnd rune
}

func (c BarConfig) SetBackets(b Brackets) BarConfig {
	c.LeftEnd = b.LeftEnd
	c.RightEnd = b.RightEnd
	return c
}

type Brackets struct {
	// LeftEnd is the default character in the left most part of the progress indicator
	LeftEnd rune
	// RightEnd is the default character in the right most part of the progress indicator
	RightEnd rune
}

func (b Brackets) Swap() Brackets {
	b.LeftEnd, b.RightEnd = b.RightEnd, b.LeftEnd
	return b
}

var (
	BracketConfigs = map[int]Brackets{
		0: {'[', ']'},
		1: {'〚', '〛'},
		2: {'【', '】'},
		3: {'〖', '〗'},

		10: {'{', '}'},
		11: {'⦃', '⦄'},

		20: {'(', ')'},
		21: {'⦅', '⦆'},
		22: {'｟', '｠'},

		30: {'<', '>'},
		31: {'❮', '❯'},
		32: {'〈', '〉'},
		33: {'《', '》'},

		40: {'〔', '〕'},
		41: {'〘', '〙'},
		42: {'﹝', '﹞'},
		43: {'⦗', '⦘'},

		50: {'|', '|'},
		51: {'▏', '▏'},
		52: {'▐', '▌'},
	}

	// BarConfigs describes predefined Bar configurations identified
	// by an integer.
	BarConfigs = map[int]BarConfig{
		0: {
			Fill:     '=',
			Head:     '>',
			Empty:    '-',
			LeftEnd:  '[',
			RightEnd: ']',
		},
		1: {
			Fill: '═',
			Head: '▷',
			// Empty:    '┄',
			Empty:    '·',
			LeftEnd:  '▕',
			RightEnd: '▏',
		},
		2: {
			Fill:     '▬',
			Head:     '▶',
			Empty:    '┄',
			LeftEnd:  '▐',
			RightEnd: '▌',
		},

		10: {
			Fill:     '▒',
			Head:     '░',
			Empty:    '░',
			LeftEnd:  '▕',
			RightEnd: '▏',
		},
		11: {
			Fill:     '▓',
			Head:     '▒',
			Empty:    '▒',
			LeftEnd:  '▕',
			RightEnd: '▏',
		},
		12: {
			Fill:     '█',
			Head:     '▒',
			Empty:    '▒',
			LeftEnd:  '▕',
			RightEnd: '▏',
		},
	}

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

	// SetBarConfig sets a complete character configuration for the Bar.
	SetBarConfig(c BarConfig) Bar

	// SetPredefined sets a predefined BarConfig. If the given index
	// is not defined, the operation does nothing.
	SetPredefined(i int) Bar

	// SetFill sets the character used to indicate the
	// completed part in the visualization.
	SetFill(rune) Bar

	// SetEmpty sets the character used to indicate the
	// pending part in the visualization.
	SetEmpty(rune) Bar

	// SetLeftEnd sets the character used to start
	// the visualization.
	SetLeftEnd(rune) Bar

	// SetRightEnd sets the character used to finish
	// the visualization.
	SetRightEnd(rune) Bar

	// SetHead sets the chacter used to indicate the head of the
	// progress bar.
	SetHead(rune) Bar

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

// _Bar represents a progress bar
type _Bar struct {
	ppi.ProgressBase[Bar]

	// total of the total  for the progress bar.
	total int

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

// NewBar returns a new progress bar
func NewBar(p Container, total int) Bar {
	b := &_Bar{
		total:  total,
		width:  Width,
		config: BarConfigs[0],
	}
	self := ppi.ProgressSelf[Bar](&_barProtected{b})
	b.ProgressBase = ppi.NewProgressBase[Bar](self, p, 1, nil)
	return b
}

// AppendCompleted appends the completion percent to the progress bar
func (b *_Bar) AppendCompleted(offset ...int) Bar {
	b.AppendFunc(func(b Element) string {
		return b.(*_Bar).CompletedPercentString()
	}, offset...)
	return b
}

// PrependCompleted prepends the precent completed to the progress bar
func (b *_Bar) PrependCompleted(offset ...int) Bar {
	b.PrependFunc(func(b Element) string {
		return b.(*_Bar).CompletedPercentString()
	}, offset...)
	return b
}

// TODO: use term width

// SetWidth sets the progress visualization width.
// The value 0 disables the visualization.
func (b *_Bar) SetWidth(n uint) Bar {
	b.width = n
	return b
}

func (b *_Bar) SetBarConfig(c BarConfig) Bar {
	b.config = c
	return b
}

func (b *_Bar) SetPredefined(i int) Bar {
	if c, ok := BarConfigs[i]; ok {
		b.config = c
	}
	return b
}

func (b *_Bar) SetBrackets(c Brackets) Bar {
	b.config = b.config.SetBackets(c)
	return b
}

func (b *_Bar) SetPredefinedBrackets(i int) Bar {
	if c, ok := BracketConfigs[i]; ok {
		b.config = b.config.SetBackets(c)
	}
	return b
}

func (b *_Bar) SetHead(c rune) Bar {
	b.config.Head = c
	return b
}

func (b *_Bar) SetEmpty(c rune) Bar {
	b.config.Empty = c
	return b
}

func (b *_Bar) SetFill(c rune) Bar {
	b.config.Fill = c
	return b
}

func (b *_Bar) SetLeftEnd(c rune) Bar {
	b.config.LeftEnd = c
	return b
}

func (b *_Bar) SetRightEnd(c rune) Bar {
	b.config.RightEnd = c
	return b
}

// Set the current count of the bar. It returns ErrMaxCurrentReached when trying n exceeds the total value. This is atomic operation and concurrency safe.
func (b *_Bar) Set(n int) bool {
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
func (b *_Bar) Incr() bool {
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

func (b *_Bar) isFinished() bool {
	b.Lock.RLock()
	defer b.Lock.RUnlock()
	return b.current == b.total
}

func (b *_Bar) incr() bool {
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

// CompletedPercentString returns the formatted string representation of the completed percent
func (b *_Bar) CompletedPercentString() string {
	return fmt.Sprintf("%3.f%%", b.CompletedPercent())
}

////////////////////////////////////////////////////////////////////////////////
