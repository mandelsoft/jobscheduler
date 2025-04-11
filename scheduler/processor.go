package scheduler

import (
	"context"
	"io"

	"github.com/mandelsoft/jobscheduler/ctxutils"
)

type schedulingContext struct {
	context.Context
	io.Writer
	job Job
}

func (c *schedulingContext) Job() Job {
	return c.job
}

func (c *schedulingContext) Scheduler() Scheduler {
	return c.job.GetScheduler()
}

///////////////////////////////////////////////////////////////////////////////

type processor struct {
	id        int
	scheduler *scheduler
}

func (p *processor) Run(ctx context.Context) {
	log.Debug("starting processor {{processor}}", "processor", p.id)
	for {
		job, err := p.scheduler.pending.Get(ctx)
		if err != nil {
			log.Debug("cancel processor {{processor}}", "processor", p.id, "error", err)
			break
		}
		if job == nil {
			log.Debug("discard processor {{processor}}", "processor", p.id)
			break
		}
		if ctxutils.IsCanceled(job.ctx) {
			log.Debug("job {{job}} was cancelled", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
			job.SetState(p.scheduler.discarded)
		} else {
			log.Debug("start job {{job}} on processor {{processor}}", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
			job.SetState(p.scheduler.running)
			job.result, job.err = job.definition.runner.Run(&schedulingContext{setJob(job.ctx, job), job.writer, job})
			job.finish()
			log.Debug("job {{job}} finished", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
		}
	}
}
