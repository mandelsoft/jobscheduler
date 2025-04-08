package scheduler

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/mandelsoft/jobscheduler/ctxutils"
	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

var jobAttr = ctxutils.NewAttribute[Job]()

func GetJob(ctx context.Context) Job {
	return jobAttr.Get(ctx)
}

func setJob(ctx context.Context, j *job) context.Context {
	return jobAttr.Set(ctx, j)
}

type job struct {
	id        string
	lock      synclog.Mutex
	scheduler *scheduler

	definition JobDefinition
	state      stateJobs
	err        error
	result     Result
	handlers   []EventHandler

	extension JobExtension
	writer    io.Writer

	wg sync.WaitGroup
}

var _ Job = (*job)(nil)

func (j *job) GetId() string {
	return j.id
}

func (j *job) String() string {
	return fmt.Sprintf("%s[%s]", j.id, j.state.State())
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

func (j *job) GetExtension(typ string) JobExtension {
	if j.extension == nil {
		return nil
	}
	return j.extension.GetExtension(typ)
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
	if jobs == j.state {
		// return
	}
	if j.state != nil {
		j.state.Remove(j)
	}

	old := j.state
	j.state = jobs
	jobs.Add(j)
	if old != nil && (old.State() == INITIAL || old.State() == WAITING || old.State() == PENDING) && jobs.State() == RUNNING {
		j.extension.Start()
	}
	j.extension.SetState(jobs.State())
	e := JobEvent{j, jobs.State()}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	// fmt.Printf("report event %s\n", e)
	go func(handlers []EventHandler) {
		// fmt.Printf("report start %s\n", e)
		for _, h := range handlers {
			h.HandleJobEvent(e)
		}
		j.scheduler.Raise(e)
		// fmt.Printf("report finished %s\n", e)
		wg.Done()
	}(slices.Clone(j.handlers))

	j.lock.Unlock()
	wg.Wait()
	switch jobs.State() {
	case DONE, DISCARDED:
		// fmt.Printf("job %s %s\n", j.id, jobs.State())
		j.extension.Close()
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

	log.Debug("schedule job", "job", j.id)

	j.wg.Add(1)
	if j.definition.discard != nil {
		js := j.definition.discard.GetState()
		if js.Valid {
			if js.Enabled {
				return j.assign(j.scheduler.discarded)
			}
		}
	}

	var jobs stateJobs
	if j.definition.trigger == nil || j.definition.trigger.IsEnabled() {
		jobs = j.scheduler.pending
	} else {
		jobs = j.scheduler.waiting
	}
	return j.assign(jobs)
}

func (j *job) Wait() {
	j.wg.Wait()
}
