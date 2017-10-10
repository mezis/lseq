package uid_test

import (
	"math/big"

	"github.com/mezis/lseq/uid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Uid", func() {
	x := uid.Generate()
	Describe("New(ToBig())", func() {
		It("is idempotent", func() {
			y := new(big.Int)
			z := uid.New(x.ToBig(y))
			Expect(z).To(Equal(x))
		})
	})
})
