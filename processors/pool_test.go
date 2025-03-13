package processors_test

import (
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/syncutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Test Environment", func() {
	var p processors.Pool
	var m processors.Mutex

	BeforeEach(func() {
		p = processors.NewDefaultPool()
		m = processors.NewMutex(p)
	})

	It("", func(ctx SpecContext) {
		MustBeSuccessful(m.Lock(ctx))
		wg := syncutils.NewWaitGroup()
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			MustBeSuccessful(m.Lock(ctx))
			m.Unlock()
			wg.Done()
		}()
		time.Sleep(100 * time.Millisecond)
		m.Unlock()
		MustBeSuccessful(wg.Wait(ctx))
		Expect("").To(Equal(""))
	}, SpecTimeout(2*time.Second))
})
