package processors_test

import (
	"github.com/mandelsoft/jobscheduler/processors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processors Test Environment", func() {
	Context("ids", func() {
		var ids processors.Ids

		BeforeEach(func() {
			ids = processors.Ids{}
		})

		It("add", func() {
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(3))
		})

		It("fill 1", func() {
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(3))

			ids.Remove(1)
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(4))
		})

		It("fill 2", func() {
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(3))

			ids.Remove(1)
			ids.Remove(2)
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(4))
		})
		It("fill middle", func() {
			Expect(ids.Next()).To(Equal(1))
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(3))

			ids.Remove(2)
			Expect(ids.Next()).To(Equal(2))
			Expect(ids.Next()).To(Equal(4))
		})

	})
})
