package lseq_test

import (
	. "github.com/mezis/lseq"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position", func() {
	Describe("Add", func() {
		It("Returns a greater position", func() {
			p0, _ := new(Position).Add(0, 0xABCD)
			p1, _ := p0.Add(0, 0xABCD)
			Expect(p0.IsBefore(p1)).To(BeTrue())
			Expect(p1.IsBefore(p0)).To(BeFalse())
		})
	})
	Describe("Interval", func() {
		It("is zero beetween empty positions", func() {
			p1 := new(Position)
			p2 := new(Position)
			Expect(p1.Interval(p2, 1)).To(Equal(int64(0)))
		})
	})
	Describe("IsBefore", func() {})
})
