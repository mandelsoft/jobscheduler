package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions/progress"
	"github.com/mandelsoft/ttyprogress"
)

func main() {
	prog := ttyprogress.For(os.Stdout)
	ext := progress.New(prog)

	sched := scheduler.New("demo1")

	sched.AddProcessor(2)
	sched.SetExtension(ext)
	sched.Run(context.Background())

	d1 := scheduler.DefineJob("job", scheduler.RunnerFunc(runner1)).
		SetExtension(progress.Define(ttyprogress.NewSpinner().
			PrependVariable(progress.VAR_JOBID).
			PrependVariable(progress.VAR_JOBSTATE).
			SetPredefined(1000).
			SetDone("").
			SetSpeed(2),
		))

	d2 := scheduler.DefineJob("job", scheduler.RunnerFunc(runner2)).
		SetExtension(progress.Define(ttyprogress.NewBar().
			PrependVariable(progress.VAR_JOBID).
			PrependVariable(progress.VAR_JOBSTATE).
			PrependCompleted().
			AppendElapsed().
			SetPredefined(1).
			SetAutoClose(false),
		))

	job1, _ := sched.Apply(d1, nil)
	job2, _ := sched.Apply(d1, nil)
	job3, _ := sched.Apply(d2, nil)

	job1.Schedule()
	job3.Schedule()
	time.Sleep(time.Duration((1000 + rand.Intn(100))) * time.Millisecond)
	job2.Schedule()

	job1.Wait()
	job2.Wait()
	job3.Wait()
	prog.Close()
	prog.Wait(nil)

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

	ext := progress.GetExtension(job)

	bar := ext.GetIndicator().(ttyprogress.Bar)
	lines := rand.Intn(10) + 10
	bar.SetTotal(lines)
	for i := 0; i < lines; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
		fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
		bar.Set(i + 1)
	}
	return nil, nil
}
