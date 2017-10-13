package lseq_test

import (
	"fmt"
	"runtime"
	"testing"

	. "github.com/mezis/lseq"
	"github.com/mezis/lseq/uid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Allocator", func() {
	Describe("Call", func() {
		It("inserts between non-contiguous positions of same length", func() {
			p1 := makePosition(21, 42)
			p2 := makePosition(21, 44)
			q := new(Position)
			a := NewAllocator()

			a.Call(q, p1, p2, 0xF00F00F0)

			Expect(q.String()).To(Equal("<21 @DEADBEEF, 43 @F00F00F0>"))
		})

		It("adds a level between contiguous positions", func() {
			p1 := makePosition(16, 30)
			p2 := makePosition(16, 31)
			q := new(Position)
			a := NewAllocator()

			a.Call(q, p1, p2, 0xF00F00F0)

			// fmt.Printf("pObs = %#v\n", pObs)
			Expect(q.DigitAt(0)).To(Equal(16))
			Expect(q.DigitAt(1)).To(Equal(30))
			Expect(q.Length()).To(Equal(3))
			Expect(q.SiteAt(2)).To(Equal(uid.Uid(0xF00F00F0)))
		})
	})
})

var pos Position

func BenchmarkPositionAllocate(b *testing.B) {
	alloc := NewAllocator()
	for _, n := range bmLengths {
		N := 1000
		l := make([]*Position, 2*N)
		for k := 0; k < N; k++ {
			p1 := genPosition(n)
			p2 := genPosition(n)
			if !p1.IsBefore(p2) {
				p1, p2 = p2, p1
			}
			if !p1.IsBefore(p2) {
				// we swapped but still not ordered - the positions are equal.
				// roll dice again.
				k--
				continue
			}
			l[2*k] = p1
			l[2*k+1] = p2
		}
		runtime.GC()

		b.Run(fmt.Sprintf("length=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for k := 0; k < b.N; k++ {
				j := k % N
				alloc.Call(&pos, l[2*j], l[2*j+1], 0xF00F00F0)
			}
		})
	}
}
