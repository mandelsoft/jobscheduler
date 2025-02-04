package uiprogress

import (
	"io"

	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type Element = ppi.Element

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc = ppi.DecoratorFunc

////////////////////////////////////////////////////////////////////////////////

type Progress interface {
	ppi.Progress
}

func New(out io.Writer) Progress {
	return ppi.New(out)
}

////////////////////////////////////////////////////////////////////////////////

type Bar interface {
	ppi.Bar
}

func NewBar(p Progress, total int) Bar {
	b := ppi.NewBar(p, total)
	return b
}

////////////////////////////////////////////////////////////////////////////////

// SpinnerTypes provides a predefined set of
// different spinner types
var SpinnerTypes = ppi.SpinnerTypes

type Spinner interface {
	ppi.Spinner
}

func NewSpinner(p Progress, set int) Spinner {
	return ppi.NewSpinner(p, set)
}

////////////////////////////////////////////////////////////////////////////////

type Steps interface {
	ppi.Steps
}

func NewSteps(p Progress, steps ...string) Steps {
	return ppi.NewSteps(p, steps...)
}

////////////////////////////////////////////////////////////////////////////////

type Text interface {
	ppi.Text
}

func NewText(p Progress, view ...int) Text {
	return ppi.NewText(p, view...)
}

////////////////////////////////////////////////////////////////////////////////

type TextSpinner interface {
	ppi.TextSpinner
}

func NewTextSpinner(p Progress, set int, view ...int) TextSpinner {
	return ppi.NewTextSpinner(p, set, view...)
}
