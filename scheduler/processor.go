package scheduler

import (
	"context"
)

type processor struct {
	id        int
	scheduler *scheduler
}

func (p *processor) run(id int, ctx context.Context) {
	p.id = id
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
