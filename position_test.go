package lseq_test

import (
	"fmt"
	"math/rand"

	. "github.com/mezis/lseq"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: change Add() to also return an error, and create a
// makePosition(digits...uint) helper here.

var _ = Describe("Position", func() {
	makePosition := func(digits ...uint) *Position {
		out := new(Position)
		for _, digit := range digits {
			out = out.Add(digit, 0xDEADBEEF)
		}
		return out
	}

	Describe("Add", func() {
		It("Returns a greater position", func() {
			p0 := makePosition(1)
			p1 := makePosition(1, 1)
			Expect(p0.IsBefore(p1)).To(BeTrue())
			Expect(p1.IsBefore(p0)).To(BeFalse())
		})
	})

	Describe("IndexAt", func() {
		p0 := makePosition()
		p1 := makePosition(23)
		p2 := makePosition(23, 42)

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
			p0 := makePosition(23)
			p1 := makePosition(23, 42)
			Expect(p1.Interval(p0, 1)).To(Equal(42))
		})

		It("is the index difference at a given depth", func() {
			p0 := makePosition(21, 21)
			p1 := makePosition(23, 42)
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
			p1 := makePosition(21, 42)
			p2 := makePosition(21, 42)
			Expect(p1.IsBefore(p2)).To(BeFalse())
			Expect(p2.IsBefore(p1)).To(BeFalse())

		})

		It("is true with LHS start-sentinel", func() {
			p1 := new(Position).Add(0, 0)
			p2 := new(Position).Add(1, 0)
			check(p1, p2)
		})

		It("is true when LHS is a prefix", func() {
			p1 := makePosition(21)
			p2 := makePosition(21, 61, 121)
			check(p1, p2)
		})

		It("is true when same lengths and lower index", func() {
			p1 := makePosition(21, 61)
			p2 := makePosition(21, 62)
			check(p1, p2)
		})

		It("is true when equal indices, LHS lower site ID", func() {
			p1 := makePosition(21).Add(61, 0)
			p2 := makePosition(21).Add(61, 1)
			check(p1, p2)
		})

		It("is true even when the RHS is shorter", func() {
			p1 := makePosition(21, 42, 1)
			p2 := makePosition(21, 43)
			check(p1, p2)
		})
	})

	Describe("Allocate", func() {
		It("inserts between non-contiguous positions of same length", func() {
			p1 := makePosition(21, 42)
			p2 := makePosition(21, 44)
			m := make(StrategyMap)

			pObs := Allocate(p1, p2, m, 0xDEADBEEF)

			fmt.Printf("pObs = %#v\n", pObs)
			Expect(pObs.IndexAt(1)).To(Equal(43))
			Expect(pObs.IndexAt(2)).To(Equal(0))
		})

		It("adds a level between contiguous positions", func() {})

		It("", func() {})

		genPosition := func() *Position {
			out := new(Position)
			for d := uint(0); d < 3; d++ {
				digit := rand.Intn(1<<(d+5) - 1)
				out = out.Add(uint(digit), 0xFF)
			}
			return out
		}

		XMeasure("runs quickly", func(b Benchmarker) {
			p1 := genPosition()
			p2 := genPosition()
			m := make(StrategyMap)
			if p2.IsBefore(p1) {
				p1, p2 = p2, p1
			}

			b.Time("runtime", func() {
				Allocate(p1, p2, m, 0xAA)
			})
		}, 1000)

	})
})
