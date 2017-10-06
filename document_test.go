package lseq_test

import (
	"github.com/mezis/lseq"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Document", func() {
	Describe("NewDocument", func() {
		It("passes", func() {
			x := lseq.NewDocument()
			Expect(x).NotTo(Equal(nil))
		})
	})
})
