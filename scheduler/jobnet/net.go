package jobnet

import (
	"fmt"
	"maps"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type NetContext struct {
	Jobs       map[string]scheduler.Job
	Conditions map[string]condition.Condition
	Payload    any
}

type Net struct {
	def  scheduler.DefaultJobDefinition
	jobs map[string]Job
}

func DefineNet(name string) Net {
	return Net{def: scheduler.DefaultJobDefinition{}.SetName(name), jobs: make(map[string]Job)}
}

func (n Net) AddJob(jobs ...Job) Net {
	n.jobs = maps.Clone(n.jobs)
	for _, j := range jobs {
		n.jobs[j.GetName()] = j
	}
	return n
}

func (n Net) Validate() error {
	_, err := newNetInfo(n)
	return err
}

func (n Net) For(payload any) (scheduler.JobDefinition, error) {
	info, err := newNetInfo(n)
	if err != nil {
		return nil, err
	}
	def := scheduler.DefineJob(n.GetName(), newRunner(payload, info), nil).
		SetExtension(n.def.GetExtension()).
		SetCondition(n.def.GetCondition()).
		SetDiscardCondition(n.def.GetDiscardCondition())
	return def, nil
}

////////////////////////////////////////////////////////////////////////////////

func (n Net) GetName() string {
	return n.def.GetName()
}

func (n Net) SetCondition(c condition.Condition) Net {
	n.def = n.def.SetCondition(c)
	return n
}

func (n Net) SetDiscardCondition(c condition.Condition) Net {
	n.def = n.def.SetDiscardCondition(c)
	return n

}

func (n Net) SetPriority(p scheduler.Priority) Net {
	n.def = n.def.SetPriority(p)
	return n
}

func (n Net) SetExtension(s scheduler.ExtensionDefinition) Net {
	n.def = n.def.SetExtension(s)
	return n
}

////////////////////////////////////////////////////////////////////////////////

type netInfo struct {
	Net
	conds   map[string]condition.Condition
	ordered []string
}

func newNetInfo(n Net) (*netInfo, error) {
	result := errors.ErrListf("inconsistent jobnet %q", n.GetName())
	info := &netInfo{
		conds: map[string]condition.Condition{},
		Net:   n,
	}
	depends := map[string]set.Set[string]{}

	for name, j := range n.jobs {
		required, err := j.validate(n.jobs)
		result.Add(err)
		depends[name] = required

		if j.trigger != nil {
			result.Add(j.trigger.Prepare(info.conds))
		}
		if j.discard != nil {
			result.Add(j.discard.Prepare(info.conds))
		}
	}

	var cycles [][]string
	info.ordered, cycles = order(depends)

	if len(cycles) > 0 {
		result.Add(fmt.Errorf("found cycles: %v", cycles))
	}

	return info, result.Result()
}

////////////////////////////////////////////////////////////////////////////////

type netRunner struct {
	payload any
	jobs    map[string]scheduler.Job
	info    *netInfo
	wg      *processors.WaitGroup
}

var _ scheduler.Runner = (*netRunner)(nil)

func newRunner(payload any, info *netInfo) *netRunner {
	return &netRunner{
		jobs:    map[string]scheduler.Job{},
		payload: payload,
		info:    info,
	}
}

func (r *netRunner) Run(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
	var gerr error

	netctx := &NetContext{
		Conditions: r.info.conds,
		Jobs:       r.jobs,
		Payload:    r.payload,
	}

	for _, n := range r.info.ordered {
		fmt.Fprintf(ctx, "creating job %q\n", n)
		d := r.info.jobs[n]
		c, err := createTrigger(d.trigger, netctx)
		if err != nil {
			gerr = errors.Wrapf(err, "condition for %q", n)
			break
		}
		dc, err := createTrigger(d.discard, netctx)
		if err != nil {
			gerr = errors.Wrapf(err, "discard condition for %q", n)
			break
		}
		job, err := ctx.Scheduler().Apply(scheduler.DefineJob(n, r.runner(d.runner.CreateRunner(netctx))).
			SetExtension(d.extension).
			SetPriority(d.priority).
			SetCondition(c).
			SetDiscardCondition(dc), ctx.Job())
		if err != nil {
			gerr = errors.Wrapf(err, "connot appy job %q", n)
			break
		}
		netctx.Jobs[n] = job
	}
	if gerr != nil {
		for _, j := range r.jobs {
			j.Cancel()
		}
		return nil, gerr
	}

	r.wg = processors.NewWaitGroup()
	for n, j := range r.jobs {
		fmt.Fprintf(ctx, "scheduling job %q\n", n)
		r.wg.Add(1)
		j.Schedule()
	}
	fmt.Fprintf(ctx, "wait for net jobs to be finished\n")
	return nil, r.wg.Wait(ctx)
}

func (r *netRunner) runner(runner scheduler.Runner) scheduler.Runner {
	return scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
		defer r.wg.Done()
		return runner.Run(ctx)
	})
}

func createTrigger(c Condition, ctx *NetContext) (condition.Condition, error) {
	if c == nil {
		return nil, nil
	}
	return c.Create(ctx)
}
