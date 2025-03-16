package processors_test

import (
	"fmt"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/jobscheduler/processors"
	"github.com/mandelsoft/jobscheduler/syncutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Limit Pool Test Environment", func() {
	var pool processors.LimitPool

	BeforeEach(func() {
		pool = processors.NewLimitPool(1)
	})

	It("limit 1", func(ctx SpecContext) {
		lock := processors.NewMutex(pool)
		MustBeSuccessful(lock.Lock(ctx))

		wg := syncutils.NewWaitGroup()
		wg.Add(10)
		now := time.Now()
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				fmt.Printf("alloc %d\n", i)
				MustBeSuccessful(pool.Alloc(ctx))
				fmt.Printf("locking %d\n", i)
				MustBeSuccessful(lock.Lock(ctx))
				fmt.Printf("locked %d\n", i)
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("unlocking %d\n", i)
				lock.Unlock()
				pool.Release()
				wg.Done()
			}()
		}

		lock.Unlock()
		MustBeSuccessful(wg.Wait(ctx))
		Expect(time.Now().Sub(now)).To(BeNumerically(">", 1000*time.Millisecond))
	}, SpecTimeout(3*time.Second))

	It("limit 2", func(ctx SpecContext) {
		lock := processors.NewMutex(pool)
		MustBeSuccessful(lock.Lock(ctx))

		pool.Inc()
		wg := syncutils.NewWaitGroup()
		wg.Add(10)
		now := time.Now()
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				fmt.Printf("alloc %d\n", i)
				MustBeSuccessful(pool.Alloc(ctx))
				fmt.Printf("locking %d\n", i)
				MustBeSuccessful(lock.Lock(ctx))
				fmt.Printf("locked %d\n", i)
				fmt.Printf("unlocking %d\n", i)
				lock.Unlock()
				time.Sleep(100 * time.Millisecond)
				pool.Release()
				wg.Done()
			}()
		}

		lock.Unlock()
		MustBeSuccessful(wg.Wait(ctx))
		Expect(time.Now().Sub(now)).To(BeNumerically(">", 500*time.Millisecond))
		Expect(time.Now().Sub(now)).To(BeNumerically("<", 1000*time.Millisecond))
	}, SpecTimeout(3*time.Second))
})
