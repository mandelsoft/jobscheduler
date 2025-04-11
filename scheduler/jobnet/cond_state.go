package jobnet

import (
	"fmt"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

func JobStateReached(name string, state State) Condition {
	return &jobState{name, string(state), func(s State) bool { return s == state }}
}

type jobState struct {
	name  string
	desc  string
	check func(state State) bool
}

func (c *jobState) Prepare(conds map[string]condition.Condition) error {
	return nil
}

func (c *jobState) Create(ctx *NetContext) (condition.Condition, error) {
	job := ctx.Jobs[c.name]
	if job == nil {
		return nil, fmt.Errorf("job %q not found for job state condition %q", c.name, c.desc)
	}
	return scheduler.JobStateReachedByFunc(job, c.check), nil
}

func (c *jobState) Validate(jobs map[string]Job) (set.Set[string], error) {
	var err error
	if _, ok := jobs[c.name]; !ok {
		err = fmt.Errorf("job %q not found for job state condition %q", c.name, c.desc)
	}
	return set.New[string](c.name), err
}

func JobFinished(name string) Condition {
	return &jobState{name, "Finished", scheduler.IsFinished}
}

func JobDone(name string) Condition {
	return &jobState{name, string(scheduler.DONE), scheduler.IsDone}
}

func JobFailed(name string) Condition {
	return &jobState{name, "Failed", scheduler.IsFailed}
}

func JobDiscarded(name string) Condition {
	return &jobState{name, string(scheduler.DISCARDED), scheduler.IsDiscarded}
}
