package uiprogress

import (
	"github.com/mandelsoft/goutils/stringutils"
)

// Steps can be used to visualize a sequence of steps.
type Steps interface {
	Bar
}

// NewSteps create a Steps progress information for a given
// list of sequential steps.
func NewSteps(p Container, steps ...string) Steps {
	steps = stringutils.AlignLeft(steps, ' ')

	return NewBar(p, len(steps)).PrependFunc(func(b Element) string {
		c := b.(Bar).Current()
		if c == 0 && !b.IsStarted() {
			return stringutils.PadRight("", len(steps[0]), ' ')
		}
		if c < len(steps) {
			return steps[c]
		}
		return stringutils.PadRight("", len(steps[0]), ' ')
	})
}
