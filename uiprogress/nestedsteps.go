package uiprogress

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/strutils"
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type NestedSteps interface {
	ppi.ProgressInterface[NestedSteps]

	SetFollowUpGap(m string) NestedSteps
	AppendCompleted() NestedSteps
	PrependCompleted() NestedSteps

	Current() Element
	Incr() Element
}

type _nestedSteps struct {
	lock sync.Mutex

	steps []NestedStep
	names []string

	group *ppi.GroupBase[NestedSteps, Bar]
	main  Steps
	cur   Element
}

type NestedStep struct {
	Name    string
	Factory func(p Container, n string) Element
}

// NewNestedSteps provides a group of step related progress indicators for a
// given set of sequential steps.
// If steptitle is set to true, the step name is reported in the group title.
// If NestedSteps.SetFinal is set to the empty string, only the progress of the
// active step is shown.
func NewNestedSteps(p Container, gap string, steptitle bool, steps ...NestedStep) NestedSteps {
	names := strutils.AlignLeft(sliceutils.Transform(steps, func(step NestedStep) string { return step.Name }), ' ')

	n := &_nestedSteps{steps: steps, names: names}
	n.group, n.main = ppi.NewGroupBase[NestedSteps, Bar](p, n, gap, func(b *ppi.GroupBase[NestedSteps, Bar]) (Bar, bool) {
		var e Bar
		if steptitle {
			e = NewSteps(b, names...)
		} else {
			e = NewBar(b, len(steps))
		}
		return e, e != nil // cannot check for nil by caller because of parametric type
	})
	return n
}

func (n *_nestedSteps) SetFollowUpGap(m string) NestedSteps {
	return n.group.SetFollowUpGap(m)
}

func (n *_nestedSteps) AppendCompleted() NestedSteps {
	n.main.AppendCompleted()
	return n
}

func (n *_nestedSteps) PrependCompleted() NestedSteps {
	n.main.PrependCompleted()
	return n
}

func (n *_nestedSteps) Start() Element {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.main.IsStarted() {
		return nil
	}
	n.main.Start()
	return n.add()
}

func (n *_nestedSteps) Current() Element {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.cur
}

func (n *_nestedSteps) add() Element {
	cur := n.main.Current()
	def := n.steps[cur]
	n.cur = def.Factory(n.group, n.names[cur])
	n.cur.Start()
	return n.cur
}

func (n *_nestedSteps) Incr() Element {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.cur != nil {
		n.cur.Close()
	}
	n.main.Incr()
	if !n.main.IsFinished() {
		return n.add()
	} else {
		n.group.Close()
	}
	return nil
}

func (n *_nestedSteps) Close() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.main.IsClosed() {
		return os.ErrClosed
	}
	if n.cur != nil {
		n.cur.Close()
	}
	n.cur = nil
	n.group.Close()
	return n.main.Close()
}

////////////////////////////////////////////////////////////////////////////////

func (n *_nestedSteps) IsStarted() bool {
	return n.main.IsStarted()
}

func (n *_nestedSteps) IsClosed() bool {
	return n.main.IsClosed()
}

func (n *_nestedSteps) Wait(ctx context.Context) error {
	return n.group.Wait(ctx)
}

func (n *_nestedSteps) TimeElapsed() time.Duration {
	return n.main.TimeElapsed()
}

func (n *_nestedSteps) TimeElapsedString() string {
	return n.main.TimeElapsedString()
}

func (n *_nestedSteps) SetFinal(m string) NestedSteps {
	n.main.SetFinal(m)
	return n
}

func (n *_nestedSteps) SetColor(color *color.Color) NestedSteps {
	n.main.SetColor(color)
	return n
}

func (n *_nestedSteps) AppendFunc(f DecoratorFunc, offset ...int) NestedSteps {
	n.main.AppendFunc(f, offset...)
	return n
}

func (n *_nestedSteps) PrependFunc(f DecoratorFunc, offset ...int) NestedSteps {
	n.main.PrependFunc(f, offset...)
	return n
}

func (n *_nestedSteps) AppendElapsed(offset ...int) NestedSteps {
	n.main.AppendElapsed(offset...)
	return n
}

func (n *_nestedSteps) PrependElapsed(offset ...int) NestedSteps {
	n.main.PrependElapsed(offset...)
	return n
}
