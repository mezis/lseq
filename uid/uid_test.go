package uid_test

import (
	"github.com/mezis/lseq/uid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("New", func() {
	x := uid.New()
	Describe("ToBig", func() {
		It("returns a value", func() {
			Expect(x.ToBig()).NotTo(Equal(0))
		})
	})
})
