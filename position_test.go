package lseq_test

import (
	. "github.com/mezis/lseq"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position", func() {
	Describe("Add", func() {
		It("Returns a greater position", func() {
			p0 := new(Position).Add(0, 0xABCD)
			p1 := p0.Add(0, 0xABCD)
			Expect(p0.IsBefore(p1)).To(BeTrue())
			Expect(p1.IsBefore(p0)).To(BeFalse())
		})
	})

	Describe("IndexAt", func() {
		p0 := new(Position)
		p1 := p0.Add(23, 0xFFFFFFFF)
		p2 := p1.Add(42, 0xEFAB)

		It("returns the added index", func() {
			Expect(p1.IndexAt(0)).To(Equal(23))
			Expect(p2.IndexAt(0)).To(Equal(23))
			Expect(p2.IndexAt(1)).To(Equal(42))
		})

		It("is zero for the zero position", func() {
			Expect(p0.IndexAt(0)).To(Equal(0))
			Expect(p0.IndexAt(1)).To(Equal(0))
		})

		It("it zero at higher depths", func() {
			Expect(p1.IndexAt(1)).To(Equal(0))
			Expect(p2.IndexAt(2)).To(Equal(0))
		})
	})

	Describe("Interval", func() {
		It("is zero beetween empty positions", func() {
			p1 := new(Position)
			p2 := new(Position)
			Expect(p1.Interval(p2, 1)).To(Equal(0))
		})

		It("is the latter index with a longer position", func() {
			p0 := new(Position).Add(23, 0xABCD)
			p1 := p0.Add(42, 0xBEEF)
			Expect(p1.Interval(p0, 1)).To(Equal(42))
		})

		It("is the index difference at a given depth", func() {
			p0 := new(Position).Add(21, 0xABCD).Add(21, 0xBEEF)
			p1 := new(Position).Add(23, 0xABCD).Add(42, 0xBEEF)
			Expect(p1.Interval(p0, 0)).To(Equal(2))
			Expect(p1.Interval(p0, 1)).To(Equal(21))
		})
	})

	Describe("IsBefore", func() {
		check := func(p1 *Position, p2 *Position) {
			Expect(p1.IsBefore(p2)).To(BeTrue())
			Expect(p2.IsBefore(p1)).To(BeFalse())
		}

		It("is false when positions are equal", func() {
			p1 := new(Position).Add(21, 0xABCD).Add(42, 0xBEEF)
			p2 := new(Position).Add(21, 0xABCD).Add(42, 0xBEEF)
			Expect(p1.IsBefore(p2)).To(BeFalse())
			Expect(p2.IsBefore(p1)).To(BeFalse())

		})

		It("is true with LHS start-sentinel", func() {
			p1 := new(Position).Add(0, 0)
			p2 := new(Position).Add(1, 0)
			check(p1, p2)
		})

		It("is true when LHS is a prefix", func() {
			p1 := new(Position).Add(21, 0xABCD)
			p2 := p1.Add(61, 0xABCD).Add(121, 0xABCD)
			check(p1, p2)
		})

		It("is true when same lengths and lower index", func() {
			p1 := new(Position).Add(21, 0).Add(61, 0)
			p2 := new(Position).Add(21, 0).Add(62, 0)
			check(p1, p2)
		})

		It("is true when equal indices, LHS lower site ID", func() {
			p1 := new(Position).Add(21, 0).Add(61, 0)
			p2 := new(Position).Add(21, 0).Add(61, 1)
			check(p1, p2)
		})
	})
})
