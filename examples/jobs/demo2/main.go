package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
		SetWidth(ttyprogress.PercentTerminalSize(20)).
		SetPredefined(1).SetBracketType(33).
		PrependVariable(progress.VAR_JOBID).
		PrependVariable(progress.VAR_JOBSTATE).
		AppendCompleted().
		AppendElapsed().
		SetMinVisualizationColumn(30).
		SetAutoClose(false),
	).HideOutputOnClose())

func main() {
	data := &Data{
		Steps: 10,
		Nested: map[string]*Data{
			"sub1": &Data{
				Steps: 5,
				Nested: map[string]*Data{
					"sub1.1": &Data{
						Steps: 10,
					},
					"sub1.2": &Data{
						Steps: 10,
					},
				},
			},
			"sub2": &Data{
				Steps: 6,
				Nested: map[string]*Data{
					"sub2.1": &Data{
						Steps: 10,
						Nested: map[string]*Data{
							"nested3": &Data{
								Steps: 5,
							},
						},
					},
					"sub2.2": &Data{
						Steps: 10,
					},
				},
			},
		},
	}

	prog := ttyprogress.For(os.Stdout)
	ext := progress.New(prog)
	_ = ext

	sched := scheduler.New("demo2")

	nop, useVis := options(os.Args[1:]...)
	if useVis {
		sched.SetExtension(ext)
	}
	sched.AddProcessor(nop)

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
	var bar ttyprogress.Bar

	ext := progress.GetExtension(job)
	if ext != nil {
		bar = ext.GetIndicator().(ttyprogress.Bar)
		bar.SetTotal(p.data.Steps + 1)
	}

	for i := 0; i < p.data.Steps; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(100))) * time.Millisecond)
		fmt.Fprintf(ctx, "job %s line %d\n", job.GetId(), i+1)
		if bar != nil {
			bar.Set(i + 1)
		}
	}

	fmt.Fprintf(ctx, "job %s waiting for nested\n", job.GetId())
	err := nested.Wait(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(ctx, "job %s gathering nested\n", job.GetId())

	// processing finished
	if p.wait != nil {
		p.wait.Done()
	}
	if bar != nil {
		bar.Set(p.data.Steps + 1)
	}
	return nil, nil
}

func Error(s string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", s)
	os.Exit(1)
}

func options(args ...string) (int, bool) {
	var err error

	p := 2
	ext := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v":
			ext = true
		case "-p":
			if i+1 >= len(args) {
				Error("number of processors missing")
			}
			p, err = strconv.Atoi(args[i+1])
			if err != nil {
				Error(err.Error())
			}
			i++
		default:
			Error(fmt.Sprintf("unknown option %q", args[i]))
		}
	}
	return p, ext
}
