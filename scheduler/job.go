package scheduler

import (
	"fmt"
	"slices"
	"sync"

	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

type job struct {
	id        string
	lock      synclog.Mutex
	scheduler *scheduler

	definition JobDefinition
	state      stateJobs
	err        error
	result     Result
	handlers   []EventHandler

	wg sync.WaitGroup
}

var _ Job = (*job)(nil)

func (j *job) GetId() string {
	return j.id
}

func (j *job) GetScheduler() Scheduler {
	return j.scheduler
}

func (j *job) GetPriority() Priority {
	return j.definition.priority
}

func (j *job) GetState() State {
	j.lock.Lock()
	defer j.lock.Unlock()

	if j.state == nil {
		return INITIAL
	}
	return j.state.State()
}

func (j *job) GetResult() (Result, error) {
	j.lock.Lock()
	defer j.lock.Unlock()

	return j.result, j.err
}

func (j *job) SetState(jobs stateJobs) {
	j.lock.Lock()
	j.setState(jobs)
}

func (j *job) setState(jobs stateJobs) {
	if j.state != nil {
		j.state.Remove(j)
	}
	jobs.Add(j)
	j.state = jobs
	e := JobEvent{j, jobs.State()}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	fmt.Printf("report event %s\n", e)
	go func(handlers []EventHandler) {
		fmt.Printf("report start %s\n", e)
		for _, h := range handlers {
			h.HandleJobEvent(e)
		}
		j.scheduler.Raise(e)
		wg.Done()
		fmt.Printf("report finished %s\n", e)
	}(slices.Clone(j.handlers))

	j.lock.Unlock()
	wg.Wait()
	switch j.state.State() {
	case DONE, DISCARDED:
		fmt.Printf("job %s done\n", j.id)
		j.wg.Done()
	}
}

func (j *job) RegisterHandler(handler EventHandler) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.handlers = append(j.handlers, handler)
}

func (j *job) UnregisterHandler(handler EventHandler) {
	j.lock.Lock()
	defer j.lock.Unlock()

	for i, h := range j.handlers {
		if h == handler {
			j.handlers = append(j.handlers[:i], j.handlers[i+1:]...)
			break
		}
	}
}

func (j *job) assign(state stateJobs) error {
	j.setState(state)
	return nil
}

func (j *job) Schedule() error {
	j.lock.Lock()

	if j.state.State() != INITIAL {
		j.lock.Unlock()
		return fmt.Errorf("already scheduled")
	}

	fmt.Printf("schedule job %s\n", j.id)

	if j.definition.discard != nil {
		js := j.definition.discard.GetState()
		if js.Valid {
			if js.Enabled {
				return j.assign(j.scheduler.discarded)
			}
		}
	}

	j.wg.Add(1)

	var jobs stateJobs
	if j.definition.trigger == nil || j.definition.trigger.IsEnabled() {
		jobs = j.scheduler.ready
	} else {
		jobs = j.scheduler.waiting
	}
	return j.assign(jobs)
}

func (j *job) Wait() {
	j.wg.Wait()
}
