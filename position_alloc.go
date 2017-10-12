package lseq

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/mezis/lseq/uid"
)

// Allocator -
// Used to allocate positions. Not thread-safe.
type Allocator struct {
	// allocator state
	m StrategyMap
	// temporary variables used during allocations, set as state to minimise
	// memory allocations and garbage collection.
	n, p, q big.Int
	lt, rt  Position
}

// NewAllocator -
// Suitable to allocate a position between two others.
func NewAllocator() *Allocator {
	out := new(Allocator)
	out.n.SetUint64(0)
	return out
}

// How many bits to left-shift `digits` by, if currently of length `a`,
// to be length `b` ?
//
// TODO: this could be computed as a difference of cum-sums of bit counts, to
// avoid the loop.
func (*Allocator) shiftBits(a uint8, b uint8) uint {
	out := uint(0)
	for l := a; l < b; l++ {
		out += uint(bitsAtDepth(l))
	}
	return out
}

// Set `pos` to be a prefix of `oth` of length `length`, paddding with
// zeroes or trimming as appropriate.
// Site identifiers are ignored.
func (alloc *Allocator) setPrefix(pos *Position, oth *Position, length uint8) {
	if oth.length > length { // trim
		pos.digits.Rsh(&oth.digits, alloc.shiftBits(length, oth.length))
	} else if oth.length < length { // pad
		pos.digits.Lsh(&oth.digits, alloc.shiftBits(oth.length, length))
	} else { // copy
		pos.digits.Set(&oth.digits)
	}
	pos.length = length
}

// Call -
// Implementation of the core LSEQ allocation algorithm. Sets `out` to a new position between the
// `left` and `right` ones, with `site` as a site identifier for new digits.
func (alloc *Allocator) Call(out *Position, left *Position, right *Position, site uid.Uid) {
	//fmt.Printf("Allocator.Call(\n\t%#v,\n\t%#v)\n", left, right)
	if debug && !left.IsBefore(right) {
		panic(fmt.Sprint("arguments not in order ", left, right))
	}

	// find a depth and prefixes with a sufficient interval
	//fmt.Printf("** finding prefixes\n")
	var interval int
	var depth uint8
	for depth = 1; depth < maxDigits; depth++ {
		//fmt.Printf("*** depth %d\n", depth)
		alloc.setPrefix(&alloc.lt, left, depth)
		alloc.setPrefix(&alloc.rt, right, depth)
		interval = alloc.rt.Interval(&alloc.lt)
		//fmt.Printf("  left  = %#v\n", &alloc.lt)
		//fmt.Printf("  right = %#v\n", &alloc.rt)
		//fmt.Printf("  interval(%d) = %d\n", depth, interval)
		if interval >= 1 {
			break
		}
	}
	if debug && depth >= maxDigits {
		panic("max depth reached")
	}
	if debug && interval < 1 {
		panic("failed to locate big enough interval")
	}

	// calculate digits for the new position
	//fmt.Println("** calculate digits")
	//fmt.Println("*** interval:", interval)
	offset := rand.Intn(min(boundary, interval)) + 1

	out.length = depth
	out.sites.SetInt64(int64(0))

	alloc.n.SetInt64(int64(offset))
	s := alloc.m.Get(depth)
	switch s {
	case boundaryLoStrategy:
		out.digits.Add(&alloc.lt.digits, &alloc.n)
	case boundaryHiStrategy:
		out.digits.Sub(&alloc.rt.digits, &alloc.n)
	default:
		panic(fmt.Sprintf("unknown strategy %#v", s))
	}
	//fmt.Printf("*** strategy[%d]: %s\n", depth, s)
	//fmt.Println("*** offset:", &alloc.n)
	//fmt.Println("*** result:", out)

	// merge site identifiers
	//fmt.Println("** interleave new indentifiers")
	for d := uint8(0); d < out.length; d++ {
		// read digits
		out.digitAt(&alloc.n, d)
		left.digitAt(&alloc.p, d)
		right.digitAt(&alloc.q, d)

		// shift and set digit
		out.sites.Lsh(&out.sites, uint(bitsAtDepth(d)))
		out.sites.Or(&out.sites, &alloc.n)

		if bigEql(&alloc.n, &alloc.p) { // use left site
			left.siteAt(&alloc.n, d)
		} else if bigEql(&alloc.n, &alloc.q) { // use right site
			right.siteAt(&alloc.n, d)
		} else { // use caller site
			site.ToBig(&alloc.n)
		}

		// shift and set site
		out.sites.Lsh(&out.sites, uid.Bits)
		out.sites.Or(&out.sites, &alloc.n)
	}

	// check and return
	if debug && !(left.IsBefore(out) && out.IsBefore(right)) {
		//fmt.Printf("left = %#v\n", left)
		//fmt.Printf("out = %#v\n", out)
		//fmt.Printf("right = %#v\n", right)
		panic("allocated position not in order")
	}
	//fmt.Println("** returning ", out)
}

func bigEql(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) == 0
}
