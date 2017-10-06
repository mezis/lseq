package lseq_test

import (
	"github.com/mezis/lseq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// "unsafe"
)

var _ = Describe("lseq", func() {
	Describe("NewDocument", func() {
		It("passes", func() {
			x := lseq.NewDocument()
			Expect(x).NotTo(Equal(nil))
		})
	})

	// Describe("Position", func() {
	// 	It("is stored with 40 bytes", func() {
	// 		x := lseq.Position{}
	// 		Expect(int(unsafe.Sizeof(x))).To(Equal(40))
	// 		Expect(int(unsafe.Sizeof(&x))).To(Equal(40))
	// 	})
	// })
})
