package jobnet

import (
	"github.com/mandelsoft/goutils/set"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Order Test Environment", func() {
	Context("order", func() {
		It("simple order", func() {
			elems := map[string]set.Set[string]{
				"A": set.New[string]("B", "C"),
				"B": nil,
				"C": set.New[string]("B", "D"),
				"D": nil,
				"E": set.New[string]("D"),
			}

			ordered, cycles := order(elems)
			Expect(cycles).To(BeNil())

			Expect(ordered).To(ContainInOrder(
				"D", "E",
			))
			Expect(ordered).To(ContainInOrder(
				"B", "A",
			))
			Expect(ordered).To(ContainInOrder(
				"D", "C", "A",
			))
		})

		It("cycle", func() {
			elems := map[string]set.Set[string]{
				"A": set.New[string]("B", "C"),
				"B": nil,
				"C": set.New[string]("B", "D"),
				"D": set.New[string]("A"),
				"E": set.New[string]("D"),
			}

			_, cycles := order(elems)
			Expect(cycles).To(Equal([][]string{[]string{"D", "A", "C", "D"}}))

		})
	})
})
