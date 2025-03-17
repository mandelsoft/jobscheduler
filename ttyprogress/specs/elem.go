package specs

import (
	"context"
	"io"
	"time"
)

// ElementInterface is the common interface of all
// elements provided by the ttyprogress package
type ElementInterface interface {
	io.Closer

	// Start records the actual start time and
	// starts the element.
	Start()

	// IsStarted reports whether element has been started.
	IsStarted() bool

	// IsClosed reports whether element has already been closed.
	IsClosed() bool

	// TimeElapsed reports the duration this element has been
	// active (time since Start or between Start and Close).
	TimeElapsed() time.Duration

	// TimeElapsedString provides a nice string representation for
	// TimeElapsed.
	TimeElapsedString() string

	// Flush emits the current output.
	Flush() error

	// Wait waits until the element is finished.
	Wait(ctx context.Context) error
}

type ElementDefinition[T any] struct {
	self  Self[T]
	final string
}

var (
	_ ElementSpecification[any] = (*ElementDefinition[any])(nil)
	_ ElementConfiguration      = (*ElementDefinition[any])(nil)
)

func NewElementDefinition[T any](self Self[T]) ElementDefinition[T] {
	return ElementDefinition[T]{self: self}
}

func (d *ElementDefinition[T]) Self() T {
	return d.self.Self()
}

func (d *ElementDefinition[T]) Dup(s Self[T]) ElementDefinition[T] {
	dup := *d
	dup.self = s
	return dup
}

func (e *ElementDefinition[T]) SetFinal(m string) T {
	e.final = m
	return e.self.Self()
}

func (e *ElementDefinition[T]) GetFinal() string {
	return e.final
}

////////////////////////////////////////////////////////////////////////////////

// TitleLineProvider is the optional interface to provide a title line configuration
// enriching the element configuration.
type TitleLineProvider interface {
	GetTitleLine() string
}

// GapProvider is the optional interface to provide an indent
// enriching the element configuration.
type GapProvider interface {
	GetGap() string
}

// FollowupGapProvider is the optional interface to provide an indent
// for additional lines
// enriching the element configuration.
type FollowupGapProvider interface {
	GetFollowUpGap() string
}

type ElementSpecification[T any] interface {
	// SetFinal sets a text message shown instead of the
	// text window after the action has been finished.
	SetFinal(string) T
}

type ElementConfiguration interface {
	// GetFinal gets a text message shown instead of the
	// text window after the action has been finished.
	GetFinal() string
}
