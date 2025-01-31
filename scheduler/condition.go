package scheduler

import (
	"fmt"
	"sync"

	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

type JobEvent struct {
	job   Job
	state State
}

func (e JobEvent) String() string {
	return fmt.Sprintf("%s:%s", e.job.GetId(), e.state)
}

func (e JobEvent) GetState() State {
	return e.state
}

func (e JobEvent) GetJob() Job {
	return e.job
}

func (e JobEvent) GetJobId() string {
	return e.job.GetId()
}

////////////////////////////////////////////////////////////////////////////////

type jobStateReached struct {
	lock sync.Mutex

	job   Job
	check func(State) bool

	reached bool
}

var _ condition.Condition = (*jobStateReached)(nil)

func JobStateReached(job Job, state State) condition.Condition {
	return &jobStateReached{job: job, check: func(s State) bool { return state == s }}
}

func JobFinished(job Job) condition.Condition {
	return &jobStateReached{job: job, check: func(s State) bool { return s == DONE || s == DISCARDED }}
}

func (t *jobStateReached) GetState() condition.State {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.reached || t.check(t.job.GetState()) {
		t.reached = true
		return condition.State{true, true, true}
	}
	return condition.State{false, false, true}
}

func (t *jobStateReached) IsEnabled() bool {
	s := t.GetState()
	return s.Enabled
}

func (t *jobStateReached) Evaluate(e condition.Event) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.reached {
		return
	}
	if j, ok := e.(JobEvent); ok {
		t.reached = j.job == t.job && t.check(j.state)
	}
}

func (t *jobStateReached) Walk(w condition.Walker) bool {
	return w.Walk(t)
}

////////////////////////////////////////////////////////////////////////////////
