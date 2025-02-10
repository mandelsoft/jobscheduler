package queue_test

import (
	"runtime"
	"sync"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/logrusl"

	"github.com/mandelsoft/jobscheduler/queue"
)

var _ = Describe("PQueue Test Environment", func() {
	logging.DefaultContext().SetBaseLogger(logrusl.Human(true).NewLogr())
	logging.DefaultContext().SetDefaultLevel(logging.DebugLevel)

	log := logging.DefaultContext().Logger(logging.NewRealm("test"))

	Context("queue", func() {
		var q queue.DiscardQueue[Element, *Element]

		BeforeEach(func() {
			q = queue.NewDiscard[Element]()
		})

		It("continue and get", func(ctx SpecContext) {

			cont := map[string]error{}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				err := q.DiscardRequest(ctx)
				log.Info("continue", "error", err)
				cont["1"] = err
				wg.Done()
			}()

			for !q.HasDiscarded() {
				runtime.Gosched()
			}
			Expect(Must(q.Get(ctx))).To(BeNil())

			wg.Wait()
			Expect(cont).To(Equal(map[string]error{"1": nil}))
		}, SpecTimeout(2*time.Second))

		It("get/block and continue", func(ctx SpecContext) {
			e1 := &Element{name: "e1"}

			found := &Found{}
			cont := map[string]error{}

			wg := &sync.WaitGroup{}
			wg.Add(2)
			go func() {
				e, err := q.Get(ctx)
				log.Info("get 1", "entry", e, "error", err)
				found.Add("1", e, err)
				wg.Done()
			}()

			for !q.HasWaiting() {
				runtime.Gosched()
			}
			log.Info("found waiting -> continue")

			rwg := &sync.WaitGroup{}
			rwg.Add(1)
			go func() {
				err := q.DiscardRequest(ctx)
				log.Info("continue", "error", err)
				cont["2"] = err
				rwg.Done()
				wg.Done()
			}()

			rwg.Wait()
			log.Info("continue request done -> continue")

			q.Add(e1)

			wg.Wait()
			Expect(found.found).To(Equal(map[string]entry{
				"1": {"", nil},
			}))
			Expect(cont).To(Equal(map[string]error{
				"2": nil,
			}))
		}, SpecTimeout(2*time.Second))
	})
})
