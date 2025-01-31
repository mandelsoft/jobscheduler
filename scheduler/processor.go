package scheduler

import (
	"context"
)

type Processor struct {
	id        int
	scheduler *scheduler
}

func (p *Processor) run(ctx context.Context) {
	sctx := SchedulingContext{
		p.scheduler,
	}
	for {
		job, err := p.scheduler.ready.Get(ctx)
		if err != nil {
			break
		}
		log.Debug("start job", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
		job.SetState(p.scheduler.running)
		job.definition.runner.Run(sctx)
		job.SetState(p.scheduler.done)
		log.Debug("job finished", "job", job.id, "processor", p.id, "scheduler", p.scheduler.name)
	}
}
