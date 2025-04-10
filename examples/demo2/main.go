package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions/progress"
	"github.com/mandelsoft/ttyprogress"
)

type Data struct {
	Steps  int
	Nested map[string]*Data
}

var jobDef = scheduler.DefineJob("main").
	SetExtension(progress.Define(ttyprogress.NewBar().
		PrependVariable(progress.VAR_JOBID).
		PrependVariable(progress.VAR_JOBSTATE).
		PrependCompleted().
		AppendElapsed().
		SetPredefined(1).
		SetAutoClose(false),
	).HideOutputOnClose())

func main() {
	data := &Data{
		Steps: 10,
		Nested: map[string]*Data{
			"sub1": &Data{
				Steps: 5,
			},
			"sub2": &Data{
				Steps: 6,
				Nested: map[string]*Data{
					"nested": &Data{
						Steps: 10,
					},
				},
			},
		},
	}

	prog := ttyprogress.For(os.Stdout)
	ext := progress.New(prog)

	sched := scheduler.New("demo1")

	sched.AddProcessor(2)
	sched.SetExtension(ext)
	sched.Run(context.Background())

	job, _ := sched.Apply(jobDef.SetRunner(NewDataProcessor(data)))

	job.Schedule()

	job.Wait()
	prog.Close()
	prog.Wait(nil)

}

type DataProcessor struct {
	data *Data
	wait *processors.WaitGroup
}

func NewDataProcessor(data *Data, wait ...*processors.WaitGroup) *DataProcessor {
	return &DataProcessor{data, general.Optional(wait...)}
}

func (p *DataProcessor) Run(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
	job := ctx.Job()

	// schedule nested jobs
	nested := processors.NewWaitGroup()
	for n, s := range p.data.Nested {
		nested.Add(1)
		proc := NewDataProcessor(s, nested)
		job, err := ctx.Scheduler().Apply(jobDef.SetName(n).SetRunner(proc), job)
		if err != nil {
			return nil, err
		}
		err = job.Schedule()
		if err != nil {
			return nil, err
		}
	}

	// do local work
	ext := progress.GetExtension(job)
	bar := ext.GetIndicator().(ttyprogress.Bar)
	bar.SetTotal(p.data.Steps + 1)
	for i := 0; i < p.data.Steps; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
		fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
		bar.Set(i + 1)
	}
	fmt.Fprintf(ctx, "job %s waiting for nested\n", job.GetId())
	err := nested.Wait(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(ctx, "job %s gathering nested\n", job.GetId())
	if p.wait != nil {
		p.wait.Done()
	}
	bar.Set(p.data.Steps + 1)
	return nil, nil
}
