package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions/buffered"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions/progress"
	"github.com/mandelsoft/jobscheduler/scheduler/jobnet"
	"github.com/mandelsoft/ttyprogress"
)

func main() {
	prog := ttyprogress.For(os.Stdout)
	ext := progress.New(prog)
	_ = ext

	sched := scheduler.New("jobnet")
	sched.SetExtension(buffered.New(os.Stdout))

	sched.AddProcessor(2)
	sched.Run(context.Background())

	njob1 := jobnet.DefineJob("first", jobnet.RunnerFunc(runner2))
	njob2 := jobnet.DefineJob("second", jobnet.RunnerFunc(runner1)).SetCondition(jobnet.DependsOn("first"))
	njob3 := njob1.SetName("third").SetCondition(jobnet.DependsOn("second"))
	njob4 := njob1.SetName("triggered").SetCondition(jobnet.Explicit("explicit"))
	netjob, _ := jobnet.DefineNet("jobnet").
		AddJob(njob1, njob2, njob3, njob4).For(10)

	job, _ := sched.ScheduleDefinition(netjob)

	job.Wait()
}

func runner1(net *jobnet.NetContext) scheduler.Runner {
	return scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
		job := ctx.Job()

		fmt.Fprintf(ctx, "job %s using paylosd %v\n", job.GetId(), net.Payload)
		cond := net.Conditions["explicit"].(*condition.ExplicitCondition)
		fmt.Fprintf(ctx, "job %s using explicit condition %p\n", job.GetId(), cond)

		lines := rand.Intn(10) + 10
		for i := 0; i < lines; i++ {
			time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
			if i == lines/2 {
				fmt.Fprintf(ctx, "job %s trigger condition\n", job.GetId())
				cond.Enable()
			}
			fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
		}
		return nil, nil
	})
}

func runner2(net *jobnet.NetContext) scheduler.Runner {
	return scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
		job := ctx.Job()

		fmt.Fprintf(ctx, "job %s using payladd %v\n", job.GetId(), net.Payload)

		lines := rand.Intn(10) + 10
		for i := 0; i < lines; i++ {
			time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
			fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
		}
		return nil, nil
	})
}
