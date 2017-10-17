package position_test

import (
	. "github.com/mezis/lseq/position"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StrategyMap", func() {
	Describe("Get", func() {
		It("generates a strategy", func() {
			var m StrategyMap
			s := m.Get(5)

			Expect(s).NotTo(Equal(UndefinedStrategy))
		})

		It("persists the strategy", func() {
			var m StrategyMap
			s := m.Get(5)

			for k := 0; k < 1000; k++ {
				Expect(m.Get(5)).To(Equal(s))
			}
		})
	})
})
