# A Job Scheduler for Go


Parallelism in Go can simply be achieved by creating new Go routines.
The problem hereby is, that it is not possible to restrict the number
of parallel executions. This module provides a job scheduler able
to execute jobs with a limited number of logical processors.

With additional synchronization elements, like monitors, mutexes or wait groups
it is possible for jobs to get blocked in a coordinated way, enabling
other jobs to continue ot start. The scheduler assures that never more than the
maximum limit of parallel execution are active.

The scheduler is provided by package `github.com/mandelsoft/jobscheduler/scheduler`.
The package `github.com/mandelsoft/jobscheduler/processors` offers
specific synchronization elements usable togetjer with the scheduler
and package `github.com/mandelsoft/jobscheduler/syncutils` provides
regular synchronization elements supporting cancellation by a `context.Context`.

```golang
import (
    "context"
    "fmt"
    "math/rand"
    "time"

    "github.com/mandelsoft/jobscheduler/scheduler"
)

func main() {
	sched := scheduler.New("demo")

	sched.AddProcessor(2)
	sched.Run(context.Background())

	d1 := scheduler.DefineJob("job1", scheduler.RunnerFunc(runner1))
	d2 := scheduler.DefineJob("job2", scheduler.RunnerFunc(runner2))

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
```

The scheduler offers an extension model, which can be used to handle
the output of the jobs. The extension `progress` provides a visualization
based on the progress indicators supported by [`github.com/mandelsoft/ttyprogress`](https://github.com/mandelsoft/ttyprogress).


<p align="center">
  <img src="examples/jobs/demo2/demo.gif" alt="Job Scheduler Demo" title="Job Scheduler Demo" />
</p>

This example can be found in [examples/jobs/demo2/main.go](examples/jobs/demo2/main.go). It works on a recursive data structure, 
processing each element with an own job. The execution is limited to 2 active 
jobs. The job for every element creates the jobs for the nested elements and
uses the synchronization operations to wait for their execution.

## Job Nets

Instead of creating single jobs one after the other, the package `scheduler/jobnet`
offers the possibility to predefine complete sets of jobs  with trigger dependencies
and execute the net as a whole.

Use `jobnet.DefineNet(name)` to create a new job net. With `jobnet.DefineJob(name, runner)`
job definitions for the net can be created. They are added to a net by 
`net.AddJob(job)`. The same definition can be added multiple times to the same net, 
but the should get a new name for each add by calling (`SetName(name)`). Otherwise the
new definition for an already configured name will replace the old one.

Special condition definitions provided by the same package can be used
to create conditions (for example dependencies using `DependsOn(...)`). Jobs are
referred here by their names used in the composed job net.

Job dependencies MUST NOT be cyclic and all job conditions used in a net must
resolvable by  the net. This can be checked with the `net.Validate()` call.

The execution of a job net is implemented by a regular job, which instantiates the
configured jobs with their defined conditions and the waits until those jobs are
finished.

A regular scheduler job definition is created with `net.For(payload)`. It is possible
to create the net for a dedicated payload information. which is passed to the runner.


A job net runner is a factory for creating a regular job runner.
It gets a `*jobnet.NetContext`, which offers access to the payload and explicit
(named) triggers created for jobs of this instance of the scheduled job net.
The factory then creates a regular job runner for the scheduler job definition,
which is then instantiated for the scheduler by the job net job.

An example can be found in [`examples/jobs/jobnet`](examples/jobs/jobnet/main.go)


<p align="center">
  <img src="examples/jobs/jobnet/demo.gif" alt="Job Net Demo" title="Job Net Demo" />
</p>
