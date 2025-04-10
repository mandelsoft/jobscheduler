package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/mandelsoft/jobscheduler/scheduler"
)

func main() {
	sched := scheduler.New("demo1")

	sched.AddProcessor(2)
	sched.Run(context.Background())

	d1 := scheduler.DefineJob("job", scheduler.RunnerFunc(runner1))
	d2 := scheduler.DefineJob("job", scheduler.RunnerFunc(runner2))

	job1, _ := sched.Apply(d1)
	job2, _ := sched.Apply(d1)
	job3, _ := sched.Apply(d2)

	job1.Schedule()
	job3.Schedule()
	time.Sleep(time.Duration((1000 + rand.Intn(100))) * time.Millisecond)
	job2.Schedule()

	job1.Wait()
	job2.Wait()
	job3.Wait()
}

func runner1(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
	job := ctx.Job()

	lines := rand.Intn(10) + 10
	for i := 0; i < lines; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
		fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
	}
	return nil, nil
}

func runner2(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
	job := ctx.Job()

	lines := rand.Intn(10) + 10
	for i := 0; i < lines; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
		fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
	}
	return nil, nil
}
