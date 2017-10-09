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
	Describe("IndexAt", func() {
		p0 := new(Position)
		p1, err := new(Position).Add(23, 0xFFFFFFFF)
		if err != nil {
			panic(err)
		}
		p2, err := p1.Add(42, 0xEFAB)
		if err != nil {
			panic(err)
		}

		It("returns the added index", func() {
			Expect(p1.IndexAt(0)).To(Equal(uint(23)))
			Expect(p2.IndexAt(0)).To(Equal(uint(23)))
			Expect(p2.IndexAt(1)).To(Equal(uint(42)))
		})

		It("is zero for the zero position", func() {
			Expect(p0.IndexAt(0)).To(Equal(uint(0)))
			Expect(p0.IndexAt(1)).To(Equal(uint(0)))
		})

		It("it zero at higher depths", func() {
			Expect(p1.IndexAt(1)).To(Equal(uint(0)))
			Expect(p2.IndexAt(2)).To(Equal(uint(0)))
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
