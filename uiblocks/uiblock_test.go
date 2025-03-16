package uiblocks_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UIBlock Test Environment", func() {
	var blocks *uiblocks.UIBlocks
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = bytes.NewBuffer(nil)
		blocks = uiblocks.New(buf)
	})

	It("assigns block", func() {
		b := uiblocks.NewBlock(3)

		Expect(b.Write([]byte("test\n"))).To(Equal(5))
		MustBeSuccessful(blocks.AddBlock(b))
		ExpectError(blocks.AddBlock(b)).To(Equal(uiblocks.ErrAlreadyAssigned))
		MustBeSuccessful(blocks.Flush())
		Expect(buf.String()).To(Equal("test\n"))
	})
})
