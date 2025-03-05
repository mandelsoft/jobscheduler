package syncutils_test

import (
	"context"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/jobscheduler/syncutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("waizgroup Test Environment", func() {
	var wg *syncutils.WaitGroup

	BeforeEach(func() {
		wg = syncutils.NewWaitGroup()
	})

	It("no empty", func(ctx SpecContext) {
		MustBeSuccessful(wg.Wait(ctx))
		Expect("").To(Equal(""))
	}, SpecTimeout(time.Second))

	It("timeout", func() {
		wg.Add(1)

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
		ExpectError(wg.Wait(ctx)).To(MatchError(context.DeadlineExceeded))
	})

	It("multiple", func(ctx SpecContext) {
		wg.Add(3)

		now := time.Now()
		go func() {
			time.Sleep(500*time.Millisecond - time.Now().Sub(now))
			wg.Done()
		}()
		go func() {
			time.Sleep(200*time.Millisecond - time.Now().Sub(now))
			wg.Done()
		}()
		go func() {
			wg.Done()
		}()
		MustBeSuccessful(wg.Wait(ctx))
		Expect(time.Now().Sub(now)).To(BeNumerically(">", 500*time.Millisecond))
	}, SpecTimeout(time.Second))
})
