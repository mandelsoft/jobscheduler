package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc = ppi.DecoratorFunc

type Element = ppi.Element

type Container = ppi.Container

type Ticker interface {
	Tick() bool
}

/*
// BaseInterface is the common interface of all
// elements provided by the uiprogress package
type BaseInterface interface {
	io.Closer

	// Start records the actual start time and
	// starts the element.
	Start() BaseInterface

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

	// Wait waits until the element is finished.
	Wait(context.Context) error
}

// ProgressInterface in the public interface of elements
// featuring a concrete progress information line.
type ProgressInterface[T any] interface {
	BaseInterface

	// SetFinal sets a text message shown instead of the
	// text window after the action has been finished.
	SetFinal(m string) T

	// SetColor sets a color for the progress indicator.
	SetColor(color *color.Color) T

	// AppendFunc adds a function providing some text appended
	// to the basic progress indicator.
	// If there are implicit settings, the offset can be used to
	// specify the index in the list of functions.
	AppendFunc(f DecoratorFunc, offset ...int) T

	// PrependFunc adds a function providing some text prepended
	// to the basic progress indicator.
	// If there are implicit settings, the offset can be used to
	// specify the index in the list of functions.
	PrependFunc(f DecoratorFunc, offset ...int) T

	// AppendElapsed appends the elapsed time of the action
	// or the duration of the action if the element is already closed.
	AppendElapsed(offset ...int) T

	// PrependElapsed appends the elapsed time of the action
	// or the duration of the action if the element is already closed.
	PrependElapsed(offset ...int) T
}

*/
