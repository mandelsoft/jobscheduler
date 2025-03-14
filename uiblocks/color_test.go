package uiblocks_test

import (
	"github.com/fatih/color"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Color Test Environment", func() {

	c := color.New(color.FgYellow, color.Bold)
	c.EnableColor()
	It("", func() {
		s := c.Sprintf("test")

		Expect(uiblocks.ColorLength([]byte(s))).To(Equal(7))
	})
})
