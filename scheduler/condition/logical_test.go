package condition_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

var _ = Describe("Logical Condition Test Environment", func() {
	Context("AND", func() {
		var cond condition.Condition
		var c1, c2 *condition.ExplicitCondition

		BeforeEach(func() {
			c1 = condition.Explicit()
			c2 = condition.Explicit()
			cond = condition.And(c1, c2)
		})

		DescribeTable("non final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, false, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, false, true}),
			Entry("one true", true, false, false, false, condition.State{false, false, false}),
			Entry("one true and valid", true, true, false, false, condition.State{false, false, false}),

			Entry("two false and valid", false, true, false, true, condition.State{false, false, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{false, false, true}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{false, false, true}),

			Entry("two true", true, false, true, false, condition.State{true, false, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, false, false}),

			Entry("two true and valid", true, true, true, true, condition.State{true, false, true}),
		)

		DescribeTable("first final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c1.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, true, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, true, true}),
			Entry("one true", true, false, false, false, condition.State{false, false, false}),
			Entry("one true and valid", true, true, false, false, condition.State{false, false, false}),

			Entry("two false and valid", false, true, false, true, condition.State{false, true, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{false, true, true}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{false, true, true}),

			Entry("two true", true, false, true, false, condition.State{true, false, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, false, false}),

			Entry("two true and valid", true, true, true, true, condition.State{true, false, true}),
		)

		DescribeTable("second final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c2.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, true, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, true, true}),
			Entry("one true", true, false, false, false, condition.State{false, true, false}),
			Entry("one true and valid", true, true, false, false, condition.State{false, true, false}),

			Entry("two false and valid", false, true, false, true, condition.State{false, true, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{false, false, true}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{false, false, true}),

			Entry("two true", true, false, true, false, condition.State{true, false, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, false, false}),

			Entry("two true and valid", true, true, true, true, condition.State{true, false, true}),
		)

		DescribeTable("both final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c1.SetFinal()
			c2.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, true, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, true, true}),
			Entry("one true", true, false, false, false, condition.State{false, true, false}),
			Entry("one true and valid", true, true, false, false, condition.State{false, true, false}),

			Entry("two false and valid", false, true, false, true, condition.State{false, true, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{false, true, true}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{false, true, true}),

			Entry("two true", true, false, true, false, condition.State{true, true, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, true, false}),

			Entry("two true and valid", true, true, true, true, condition.State{true, true, true}),
		)
	})

	////////////////////////////////////////////////////////////////////////////

	Context("OR", func() {
		var cond condition.Condition
		var c1, c2 *condition.ExplicitCondition

		BeforeEach(func() {
			c1 = condition.Explicit()
			c2 = condition.Explicit()
			cond = condition.Or(c1, c2)
		})

		DescribeTable("non final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, false, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, false, false}),
			Entry("one true", true, false, false, false, condition.State{true, false, false}),
			Entry("one true and valid", true, true, false, false, condition.State{true, false, true}),

			Entry("two false and valid", false, true, false, true, condition.State{false, false, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{true, false, false}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{true, false, true}),

			Entry("two true", true, false, true, false, condition.State{true, false, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, false, true}),

			Entry("two true and valid", true, true, true, true, condition.State{true, false, true}),
		)

		DescribeTable("first final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c1.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, false, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, false, false}),
			Entry("one true", true, false, false, false, condition.State{true, true, false}),
			Entry("one true and valid", true, true, false, false, condition.State{true, true, true}),

			Entry("two false and valid", false, true, false, true, condition.State{false, false, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{true, false, false}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{true, false, true}),

			Entry("two true", true, false, true, false, condition.State{true, true, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, true, true}),

			Entry("two true and valid", true, true, true, true, condition.State{true, true, true}),
		)

		DescribeTable("second final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c2.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, false, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, false, false}),
			Entry("one true", true, false, false, false, condition.State{true, false, false}),
			Entry("one true and valid", true, true, false, false, condition.State{true, false, true}),

			Entry("two false and valid", false, true, false, true, condition.State{false, false, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{true, true, false}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{true, true, true}),

			Entry("two true", true, false, true, false, condition.State{true, true, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, true, true}),

			Entry("two true and valid", true, true, true, true, condition.State{true, true, true}),
		)

		DescribeTable("both final", func(e1, v1, e2, v2 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if v2 {
				c2.SetValid()
			}
			c2.SetEnabled(e2)

			c1.SetFinal()
			c2.SetFinal()
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, false, condition.State{false, true, false}),
			Entry("two false, one valid", false, true, false, false, condition.State{false, true, false}),
			Entry("one true", true, false, false, false, condition.State{true, true, false}),
			Entry("one true and valid", true, true, false, false, condition.State{true, true, true}),

			Entry("two false and valid", false, true, false, true, condition.State{false, true, true}),
			Entry("one false and valid, other true", false, true, true, false, condition.State{true, true, false}),
			Entry("one false and valid, other true and valid", false, true, true, true, condition.State{true, true, true}),

			Entry("two true", true, false, true, false, condition.State{true, true, false}),
			Entry("two true, one valid", true, false, true, true, condition.State{true, true, true}),

			Entry("two true and valid", true, true, true, true, condition.State{true, true, true}),
		)
	})

	////////////////////////////////////////////////////////////////////////////

	Context("NOT", func() {
		var cond condition.Condition
		var c1 *condition.ExplicitCondition

		BeforeEach(func() {
			c1 = condition.Explicit()
			cond = condition.Not(c1)
		})

		DescribeTable("table", func(e1, v1, f1 bool, r condition.State) {
			if v1 {
				c1.SetValid()
			}
			c1.SetEnabled(e1)

			if f1 {
				c1.SetFinal()
			}
			Expect(cond.GetState()).To(Equal(r))
		},
			Entry("initial", false, false, false, condition.State{true, false, false}),
			Entry("final", false, false, true, condition.State{true, true, false}),
			Entry("valid", false, true, false, condition.State{true, false, true}),
			Entry("valid and final", false, true, true, condition.State{true, true, true}),

			Entry("true", true, false, false, condition.State{false, false, false}),
			Entry("true and final", true, false, true, condition.State{false, true, false}),
			Entry("true and valid", true, true, false, condition.State{false, false, true}),
			Entry("true, valid and final", true, true, true, condition.State{false, true, true}),
		)
	})
})
