package syncutils_test

import (
	"context"
	"time"

	"github.com/mandelsoft/jobscheduler/syncutils"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Select Test Environment", func() {
	It("mutex", func(ctx SpecContext) {
		lock := syncutils.NewMutex()
		wg := syncutils.NewWaitGroup()
		wg.Add(1)
		lock.Lock(ctx)

		go func() {
			defer GinkgoRecover()
			select {
			case <-ctx.Done():
			case err, ok := <-syncutils.Select(lock, nil):
				MustBeSuccessful(err)
				Expect(ok).To(BeTrue())
			}
			wg.Done()
		}()
		time.Sleep(500 * time.Millisecond)
		lock.Unlock()

		MustBeSuccessful(wg.Wait(ctx))
		Expect(lock.TryLock()).To(BeFalse())

	}, SpecTimeout(2*time.Second))

	It("mutex timeout", func(ctx SpecContext) {
		lock := syncutils.NewMutex()
		wg := syncutils.NewWaitGroup()
		wg.Add(1)
		lock.Lock(ctx)

		go func() {
			defer GinkgoRecover()
			cctx, _ := context.WithTimeout(ctx, 100*time.Millisecond)
			select {
			case <-ctx.Done():
			case err, ok := <-syncutils.Select(lock, cctx):
				ExpectError(err).To(MatchError(context.DeadlineExceeded))
				Expect(ok).To(BeTrue())
			}
			wg.Done()
		}()
		time.Sleep(500 * time.Millisecond)
		lock.Unlock()

		MustBeSuccessful(wg.Wait(ctx))
		Expect(lock.TryLock()).To(BeTrue())

	}, SpecTimeout(2*time.Second))
})
