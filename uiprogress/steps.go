package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/strutils"
)

func NewSteps(p *Progress, steps ...string) *Bar {
	steps = strutils.AlignLeft(steps, ' ')

	return NewBar(p, len(steps)).PrependFunc(func(b Element) string {
		c := b.(*Bar).Current()
		if c == 0 {
			return strutils.PadRight("", len(steps[0]), ' ')
		}
		return steps[c-1]
	})
}
