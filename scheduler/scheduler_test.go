package scheduler_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/scheduler"
)

type JobHandler struct {
	lock   sync.Mutex
	events []string
}

func (h *JobHandler) HandleJobEvent(e scheduler.JobEvent) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.events = append(h.events, fmt.Sprintf("%s:%s", e.GetJobId(), e.GetState()))
}

func EVTs(name string, states ...scheduler.State) []string {
	var r []string
	for _, s := range states {
		r = append(r, EVT(name, s))
	}
	return r
}

func EVT(name string, state scheduler.State) string {
	return fmt.Sprintf("%s:%s", name, state)
}

var _ = Describe("Scheduler Test Environment", func() {
	// logging.DefaultContext().SetBaseLogger(logrusl.Human(true).NewLogr())
	// logging.DefaultContext().SetDefaultLevel(logging.TraceLevel)

	// log := logging.DefaultContext().Logger(logging.NewRealm("test"))

	var sched scheduler.Scheduler

	BeforeEach(func() {
		sched = scheduler.New()
	})

	AfterEach(func() {
		sched.Cancel()
		sched.Wait()
	})

	Context("single processor", func() {
		BeforeEach(func() {
			sched.AddProcessor()
			sched.Run(nil)
		})

		It("processes job", func() {
			id := "test[1]"
			handler := &JobHandler{}

			def := scheduler.DefineJob("test",
				scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
					return nil, nil
				}))
			job := Must(sched.Apply(def, nil))
			job.RegisterHandler(handler)
			MustBeSuccessful(job.Schedule())
			job.Wait()

			Expect(handler.events).To(Equal(EVTs(id, scheduler.PENDING, scheduler.RUNNING, scheduler.DONE)))
		})

		It("processes sequence", func() {
			id1 := "test[1]"
			id2 := "test[2]"
			id3 := "test[3]"
			handler := &JobHandler{}

			def := scheduler.DefineJob("test",
				scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
					return nil, nil
				})).AddHandler(handler)
			job1 := Must(sched.Apply(def, nil))

			job2 := Must(sched.Apply(def.SetCondition(scheduler.DependsOn(job1)), nil))
			job3 := Must(sched.Apply(def.SetCondition(scheduler.DependsOn(job2)), nil))

			MustBeSuccessful(job3.Schedule())
			MustBeSuccessful(job2.Schedule())
			MustBeSuccessful(job1.Schedule())
			job3.Wait()

			Expect(handler.events).To(Equal([]string{
				EVT(id1, scheduler.INITIAL),
				EVT(id2, scheduler.INITIAL),
				EVT(id3, scheduler.INITIAL),
				EVT(id3, scheduler.WAITING),
				EVT(id2, scheduler.WAITING),
				EVT(id1, scheduler.PENDING),
				EVT(id1, scheduler.RUNNING),
				EVT(id1, scheduler.DONE),
				EVT(id2, scheduler.PENDING),
				EVT(id2, scheduler.RUNNING),
				EVT(id2, scheduler.DONE),
				EVT(id3, scheduler.PENDING),
				EVT(id3, scheduler.RUNNING),
				EVT(id3, scheduler.DONE),
			}))
		})
	})

	Context("multiple processors", func() {
		BeforeEach(func() {
			sched.AddProcessor()
			sched.AddProcessor()
			sched.Run(nil)
		})

		It("processes sequence", func() {
			id1 := "test[1]"
			id2 := "test[2]"
			id3 := "test[3]"
			handler := &JobHandler{}

			def := scheduler.DefineJob("test",
				scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
					return nil, nil
				})).AddHandler(handler)

			job1 := Must(sched.Apply(def, nil))
			job2 := Must(sched.Apply(def.SetCondition(scheduler.DependsOn(job1)), nil))
			job3 := Must(sched.Apply(def.SetCondition(scheduler.DependsOn(job2)), nil))

			MustBeSuccessful(job3.Schedule())
			MustBeSuccessful(job2.Schedule())
			MustBeSuccessful(job1.Schedule())
			job3.Wait()

			Expect(handler.events).To(Equal([]string{
				EVT(id1, scheduler.INITIAL),
				EVT(id2, scheduler.INITIAL),
				EVT(id3, scheduler.INITIAL),
				EVT(id3, scheduler.WAITING),
				EVT(id2, scheduler.WAITING),
				EVT(id1, scheduler.PENDING),
				EVT(id1, scheduler.RUNNING),
				EVT(id1, scheduler.DONE),
				EVT(id2, scheduler.PENDING),
				EVT(id2, scheduler.RUNNING),
				EVT(id2, scheduler.DONE),
				EVT(id3, scheduler.PENDING),
				EVT(id3, scheduler.RUNNING),
				EVT(id3, scheduler.DONE),
			}))
		})
	})

	Context("sync operations", func() {
		var barrier *Barrier

		BeforeEach(func() {
			barrier = NewBarrier(sched, 3)
			sched.AddProcessor()
			sched.Run(nil)
		})

		It("processes sequence", func() {
			id1 := "test[1]"
			id2 := "test[2]"
			id3 := "test[3]"

			// logging.DefaultContext().SetDefaultLevel(logging.TraceLevel)
			handler := &JobHandler{}

			def := scheduler.DefineJob("test",
				scheduler.RunnerFunc(func(ctx scheduler.SchedulingContext) (scheduler.Result, error) {
					fmt.Printf("running %s\n", scheduler.GetJob(ctx))
					barrier.Overcome(ctx)
					fmt.Printf("finishing %s\n", scheduler.GetJob(ctx))
					return nil, nil
				})).AddHandler(handler)
			job1 := Must(sched.Apply(def, nil))
			job2 := Must(sched.Apply(def, nil))
			job3 := Must(sched.Apply(def, nil))

			MustBeSuccessful(job3.Schedule())
			MustBeSuccessful(job2.Schedule())
			MustBeSuccessful(job1.Schedule())
			job3.Wait()
			job2.Wait()
			job1.Wait()

			Expect(handler.events).To(ContainInOrder(
				EVT(id3, scheduler.BLOCKED),
				EVT(id3, scheduler.READY),
				EVT(id3, scheduler.RUNNING),
				EVT(id3, scheduler.DONE),
			))
			Expect(handler.events).To(ContainInOrder(
				EVT(id3, scheduler.BLOCKED),
				EVT(id2, scheduler.BLOCKED),
				EVT(id2, scheduler.READY),
				EVT(id2, scheduler.RUNNING),
				EVT(id2, scheduler.DONE),
			))
			Expect(handler.events).To(ContainInOrder(
				EVT(id1, scheduler.RUNNING),
				EVT(id1, scheduler.DONE),
			))
		})
	})
})

type Barrier struct {
	monitor   processors.Monitor
	threshold int
	count     int
}

func NewBarrier(pool processors.PoolProvider, threshold int) *Barrier {
	return &Barrier{
		monitor:   processors.NewMonitor(),
		threshold: threshold,
	}
}

func (b *Barrier) Overcome(ctx context.Context) error {
	b.monitor.Lock()
	defer b.monitor.Unlock()

	b.count++
	if b.count >= b.threshold {
		b.monitor.SignalAll()
		return nil
	}
	return b.monitor.Wait(ctx)
}
