package syncutils_test

import (
	"context"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/jobscheduler/syncutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RWMutex Test Environment", func() {
	var lock syncutils.RWMutex
	BeforeEach(func() {
		lock = syncutils.NewRWMutex()
	})

	Context("write", func() {
		It("trylock", func() {
			Expect(lock.TryLock()).To(BeTrue())
			Expect(lock.TryLock()).To(BeFalse())
			lock.Unlock()
			Expect(lock.TryLock()).To(BeTrue())
			Expect(lock.TryLock()).To(BeFalse())
		})

		It("lock", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))
			Expect(lock.TryLock()).To(BeFalse())
			lock.Unlock()
			Expect(lock.TryLock()).To(BeTrue())
		}, SpecTimeout(2*time.Second))

		It("locked", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))
			cctx, _ := context.WithTimeout(ctx, 1*time.Second)
			ExpectError(lock.Lock(cctx)).To(MatchError(context.DeadlineExceeded))
		}, SpecTimeout(2*time.Second))

		It("locked/unlock", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))
			go func() {
				time.Sleep(500 * time.Millisecond)
				lock.Unlock()
			}()
			Expect(lock.TryLock()).To(BeFalse())
			MustBeSuccessful(lock.Lock(ctx))
		}, SpecTimeout(2*time.Second))

		It("sequence", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))

			wg := syncutils.WaitGroup{}
			wg.Add(3)
			go func() {
				MustBeSuccessful(lock.Lock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.Unlock()
				wg.Done()
			}()
			go func() {
				MustBeSuccessful(lock.Lock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.Unlock()
				wg.Done()
			}()
			go func() {
				MustBeSuccessful(lock.Lock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.Unlock()
				wg.Done()
			}()
			Expect(lock.TryLock()).To(BeFalse())
			lock.Unlock()
			wg.Wait(ctx)
		}, SpecTimeout(2*time.Second))
	})

	Context("read", func() {
		It("trylock", func() {
			Expect(lock.TryRLock()).To(BeTrue())
			Expect(lock.TryRLock()).To(BeTrue())
			Expect(lock.TryLock()).To(BeFalse())
			lock.RUnlock()
			Expect(lock.TryLock()).To(BeFalse())
			lock.RUnlock()
			Expect(lock.TryLock()).To(BeTrue())
			Expect(lock.TryRLock()).To(BeFalse())
			lock.Unlock()
			Expect(lock.TryRLock()).To(BeTrue())
		})

		It("lock", func(ctx SpecContext) {
			MustBeSuccessful(lock.RLock(ctx))
			Expect(lock.TryLock()).To(BeFalse())
			lock.RUnlock()
			Expect(lock.TryLock()).To(BeTrue())
		}, SpecTimeout(2*time.Second))

		It("rlock", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))
			Expect(lock.TryRLock()).To(BeFalse())
			lock.Unlock()
			Expect(lock.TryRLock()).To(BeTrue())
		}, SpecTimeout(2*time.Second))

		It("locked/unlock", func(ctx SpecContext) {
			MustBeSuccessful(lock.Lock(ctx))
			go func() {
				time.Sleep(500 * time.Millisecond)
				lock.Unlock()
			}()
			Expect(lock.TryRLock()).To(BeFalse())
			MustBeSuccessful(lock.RLock(ctx))
		}, SpecTimeout(2*time.Second))

		It("sequence", func(ctx SpecContext) {
			MustBeSuccessful(lock.RLock(ctx))

			wg := syncutils.WaitGroup{}
			wg.Add(3)
			go func() {
				MustBeSuccessful(lock.RLock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.RUnlock()
				wg.Done()
			}()
			go func() {
				MustBeSuccessful(lock.RLock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.RUnlock()
				wg.Done()
			}()
			go func() {
				MustBeSuccessful(lock.RLock(ctx))
				time.Sleep(500 * time.Millisecond)
				lock.RUnlock()
				wg.Done()
			}()
			Expect(lock.TryLock()).To(BeFalse())
			lock.RUnlock()
			wg.Wait(ctx)
			Expect(lock.TryLock()).To(BeTrue())
		}, SpecTimeout(2*time.Second))
	})
})
