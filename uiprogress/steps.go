package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/strutils"
)

// Steps can be used to visualize a sequence of steps.
type Steps interface {
	Bar
}

// NewSteps create a Steps progress information for a given
// list of sequential steps.
func NewSteps(p Progress, steps ...string) Steps {
	steps = strutils.AlignLeft(steps, ' ')

	return NewBar(p, len(steps)).PrependFunc(func(b Element) string {
		c := b.(Bar).Current()
		if c == 0 {
			return strutils.PadRight("", len(steps[0]), ' ')
		}
		return steps[c-1]
	})
}
