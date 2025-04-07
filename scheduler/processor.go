package scheduler

import (
	"context"
)

type processor struct {
	id        int
	scheduler *scheduler
}

func (p *processor) Run(ctx context.Context) {
	sctx := ctx
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
		log.Debug("start job {{job}} on processor {{processor}}", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
		job.SetState(p.scheduler.running)
		job.definition.runner.Run(setJob(sctx, job))
		job.SetState(p.scheduler.done)
		log.Debug("job {{job}} finished", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
	}
}
