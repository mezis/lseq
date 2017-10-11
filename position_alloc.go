package lseq

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"

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

func NewAllocator() *Allocator {
	out := new(Allocator)
	out.m = make(StrategyMap)
	out.n.SetUint64(0)
	return out
}

// Allocate -
// Implementation of the core LSEQ algorithm. Return a new position between the
// "left" and "right" ones.
func (alloc *Allocator) Call(out *Position, left *Position, right *Position, site uid.Uid) {
	//logger.Printf("Allocator.Call(%#v, %#v)\n", left, right)
	if debug && !left.IsBefore(right) {
		panic(fmt.Sprint("arguments not in order ", left, right))
	}

	// how many bits to left-shift `digits` by, if currently of length `a`,
	// to be length `b` ?
	shiftBits := func(a uint8, b uint8) uint {
		out := uint(0)
		for l := a; l < b; l++ {
			out += uint(bitsAtDepth(l))
		}
		return out
	}

	// Set `pos` to be a prefix of `oth` of length `length`, paddding with
	// zeroes or trimming as appropriate.
	// Site identifiers are ignored.
	setPrefix := func(pos *Position, oth *Position, length uint8) {
		if oth.length > length { // trim
			pos.digits.Rsh(&oth.digits, shiftBits(length, oth.length))
		} else if oth.length < length { // pad
			pos.digits.Lsh(&oth.digits, shiftBits(oth.length, length))
		} else { // copy
			pos.digits.Set(&oth.digits)
		}
		pos.length = length
	}

	// find a depth and prefixes with a sufficient interval
	//logger.Printf("** finding prefixes\n")
	var interval int
	var depth uint8
	for depth = 1; depth < maxDigits; depth++ {
		//logger.Printf("*** depth %d\n", depth)
		setPrefix(&alloc.lt, left, depth)
		setPrefix(&alloc.rt, right, depth)
		interval = alloc.rt.Interval(&alloc.lt)
		//logger.Printf("  left  = %#v\n", &alloc.lt)
		//logger.Printf("  right = %#v\n", &alloc.rt)
		//logger.Printf("  interval(%d) = %d\n", depth, interval)
		if interval >= 1 {
			break
		}
	}
	if debug && depth >= maxDigits {
		panic("max depth reached")
	}

	// calculate digits for the new position
	//logger.Println("** calculate digits")
	offset := rand.Intn(min(boundary, interval)) + 1

	out.length = depth
	out.sites.SetInt64(int64(0))

	alloc.n.SetInt64(int64(offset))
	s := alloc.m.get(depth)
	switch s {
	case boundaryLoStrategy:
		out.digits.Add(&alloc.lt.digits, &alloc.n)
	case boundaryHiStrategy:
		out.digits.Sub(&alloc.rt.digits, &alloc.n)
	default:
		panic(fmt.Sprintf("unknown strategy %#v", s))
	}
	//logger.Println("*** strategy:", s)
	//logger.Println("*** offset:", &alloc.n)
	//logger.Println("*** result:", out)

	// merge site identifiers
	//logger.Println("** interleave new indentifiers")
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
		//logger.Printf("left = %#v\n", left)
		//logger.Printf("out = %#v\n", out)
		//logger.Printf("right = %#v\n", right)
		panic("allocated position not in order")
	}
	//logger.Println("** returning ", out)
}

func bigEql(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) == 0
}

var logger = log.New(os.Stderr, "[lseq.Allocator]", log.Lmicroseconds)
