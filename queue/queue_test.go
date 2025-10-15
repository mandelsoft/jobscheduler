package queue_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/logrusl"

	"github.com/mandelsoft/jobscheduler/queue"
)

type Element struct {
	name string
	prio queue.Priority
}

func (e *Element) GetPriority() queue.Priority {
	return e.prio
}

func (e *Element) String() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("%s[%d]", e.name, e.prio)
}

type entry struct {
	name string
	err  error
}

type Found struct {
	lock  sync.Mutex
	found map[string]entry
}

func (f *Found) Add(key string, e *Element, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.found == nil {
		f.found = map[string]entry{}
	}
	if e != nil {
		f.found[key] = entry{e.name, err}
	} else {
		f.found[key] = entry{"", err}
	}
}

var _ = Describe("Queue Test Environment", func() {
	logging.DefaultContext().SetBaseLogger(logrusl.Human(true).NewLogr())
	logging.DefaultContext().SetDefaultLevel(logging.DebugLevel)

	log := logging.DefaultContext().Logger(logging.NewRealm("test"))

	Context("queue", func() {
		var q queue.Queue[Element, *Element]

		BeforeEach(func() {
			q = queue.New[Element]()
		})

		It("add/get", func(ctx SpecContext) {
			e1 := &Element{name: "e1"}
			e2 := &Element{name: "e2"}
			e3 := &Element{name: "e3"}

			q.Add(e1)
			q.Add(e2)

			Expect(Must(q.Get(ctx)).name).To(Equal("e1"))
			q.Add(e3)
			Expect(Must(q.Get(ctx)).name).To(Equal("e2"))
			Expect(Must(q.Get(ctx)).name).To(Equal("e3"))
		}, SpecTimeout(2*time.Second))

		It("block", func(ctx SpecContext) {
			e1 := &Element{name: "e1"}
			e2 := &Element{name: "e2"}

			found := &Found{}

			wg := &sync.WaitGroup{}
			wg.Add(2)

			bwq1 := &sync.WaitGroup{}
			bwq1.Add(1)
			go func() {
				defer GinkgoRecover()

				bwq1.Done()
				log.Info("get 1 started")
				e, err := q.Get(ctx)
				log.Info("get 1", "element", e, "error", err)
				found.Add("1", e, err)
				wg.Done()
			}()

			bwq2 := &sync.WaitGroup{}
			bwq2.Add(1)
			go func() {
				defer GinkgoRecover()

				bwq1.Wait()
				bwq2.Done()
				log.Info("get 2 started")

				e, err := q.Get(ctx)
				log.Info("get 2", "element", e, "error", err)
				found.Add("2", e, err)
				wg.Done()
			}()

			go func() {
				bwq2.Wait()
				log.Info("add 1")
				q.Add(e1)
				log.Info("add 2")
				q.Add(e2)
				log.Info("add done")
			}()

			wg.Wait()
			Expect(found.found).To(Equal(map[string]entry{
				"1": {"e1", nil},
				"2": {"e2", nil},
			}))
		}, SpecTimeout(2*time.Second))

		It("timeout", func(ctx SpecContext) {
			e1 := &Element{name: "e1"}
			e2 := &Element{name: "e2"}

			found := &Found{}

			wg := &sync.WaitGroup{}
			wg.Add(2)

			bwq1 := &sync.WaitGroup{}
			bwq1.Add(1)
			go func() {
				defer GinkgoRecover()

				log.Info("get 1 started")
				bwq1.Done()
				e, err := q.Get(ctx)
				log.Info("get 1", "element", e, "error", err)
				found.Add("1", e, err)
				wg.Done()
			}()

			bwq2 := &sync.WaitGroup{}
			bwq2.Add(1)
			go func() {
				defer GinkgoRecover()

				bwq1.Wait()
				log.Info("get 2 started")
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				bwq2.Done()

				e, err := q.Get(ctx)
				log.Info("get 2", "element", e, "error", err)
				found.Add("2", e, err)
				wg.Done()
			}()

			go func() {
				bwq2.Wait()
				time.Sleep(2 * time.Second)
				log.Info("add 1")
				q.Add(e1)
				time.Sleep(2 * time.Second)
				log.Info("add 2")
				q.Add(e2)
				log.Info("add done")
			}()

			wg.Wait()

			Expect(found.found).To(Equal(map[string]entry{
				"1": {"e1", nil},
				"2": {"", context.DeadlineExceeded},
			}))
		}, SpecTimeout(10*time.Second))
	})
})
