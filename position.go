package lseq

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/mezis/lseq/uid"
)

// Position - An immutable position in a document
type Position struct {
	length uint8
	// this "value" is interpreted as a list of 2 * "length" integers.
	// integer 2k has (rootBits+k) bits and is an index;
	// integer 2k+1 has 64 bits and is a "site identifier" (aka. author
	// identifier).
	// each pair forms an "identifier".
	value big.Int
}

// Number of bits used for the first (most significant) digit, ie. the root of
// the tree.
const rootBits uint = 5

// Maximum tree depth (ie. position length); results in 31-bit digits at the
// deepest level.
const maxDigits = 26

// How many free identifiers to leave before or after the first allocation at a
// new tree depth.
const boundary = 10

// The maximum digit value at a given tree depth.
// Note that is is also usable as a bitmask.
func maxDigitAtDepth(depth uint8) uint {
	return 1<<uint(bitsAtDepth(depth)) - 1
}

// Number of bits for indices a the given tree depth.
func bitsAtDepth(depth uint8) uint8 {
	return uint8(rootBits) + depth
}

// Return a position of length at least "length", padded with zeros on the right
// as necessary.
//
// Return the receiver if its length is already "length" or more.
func (pos *Position) padTo(length uint8) *Position {
	if pos.length >= length {
		return pos
	}

	out := new(Position)
	out.length = pos.length

	var shiftBy uint // = 0
	for out.length < length {
		shiftBy += rootBits + uint(out.length) + uid.Bits
		out.length++
	}
	out.value.Lsh(&pos.value, shiftBy)
	return out
}

// Return a position of length at most "length", removing trailing identifiers
// as necessary, and padding with zeroes otherwise.
//
// Return the receiver if already "length" or shorter.
func (pos *Position) trimTo(length uint8) *Position {
	if pos.length <= length {
		return pos
	}

	out := new(Position)
	out.length = pos.length

	var shiftBy uint // = 0
	for out.length > length {
		shiftBy += rootBits + uint(out.length-1) + uid.Bits
		out.length--
	}
	out.value.Rsh(&pos.value, shiftBy)
	return out
}

// Return a position of length exactly "length", padding or trimming as
// appropriate.
//
// Return the receiver if already the right length.
func (pos *Position) prefix(length uint8) *Position {
	return pos.trimTo(length).padTo(length)
}

// IsBefore -
// Return true iff "pos" is before "oth" in the partial order defined by Logoot.
//
// In practice, this is the lexicographical order on the list of (identifier,
// site) pairs, implemented as integer comparison after padding.
func (pos *Position) IsBefore(oth *Position) bool {
	pos = pos.padTo(oth.length)
	oth = oth.padTo(pos.length)
	return pos.value.Cmp(&oth.value) < 0
}

func (pos *Position) equals(oth *Position) bool {
	return pos.length == oth.length && pos.value.Cmp(&oth.value) == 0
}

// Add an identifier (index and site identifier) to the position,
// returning the new position
func (pos *Position) Add(digit uint, site uid.Uid) *Position {
	if pos.length >= maxDigits {
		return nil // max position length reached
	}

	digitBits := uint(bitsAtDepth(pos.length))
	if digit < 0 || digit >= 1<<digitBits {
		return nil // bad index value"
	}

	out := new(Position)
	out.length = pos.length + 1
	out.value.Lsh(&pos.value, digitBits)
	out.value.Or(&out.value, new(big.Int).SetUint64(uint64(digit)))
	out.value.Lsh(&out.value, uid.Bits)
	out.value.Or(&out.value, site.ToBig())

	return out
}

// DigitAt -
// Return the value of the `depth`s most significant digit.
func (pos *Position) DigitAt(depth uint8) int {
	if pos.length <= depth {
		return 0
	}

	shiftBy := uint(0)
	for d := pos.length - 1; d > depth; d-- {
		shiftBy += uint(bitsAtDepth(uint8(d)))
	}
	shiftBy += uid.Bits * uint(pos.length-depth)

	mask := new(big.Int).SetUint64(uint64(maxDigitAtDepth(depth)))
	val := new(big.Int).Rsh(&pos.value, shiftBy)
	val.And(val, mask)
	return int(val.Int64())
}

// Interval -
// Return the distance, in number of free identifiers, between "pos" and "oth"
// at the given depth.
// It is expected that "pos" ≺ "oth", and that both positions share a common
// prefix of length "depth",
func (pos *Position) Interval(oth *Position, depth uint8) int {
	return int(pos.DigitAt(depth)) - int(oth.DigitAt(depth))
}

// Allocate -
// Implementation of the core LSEQ algorithm. Return a new position between the
// "left" and "right" ones.
func Allocate(left *Position, right *Position, m StrategyMap, site uid.Uid) *Position {
	fmt.Printf("Allocate(%#v, %#v)\n", left, right)
	if debug && !left.IsBefore(right) {
		panic("arguments not in order")
	}

	interval := 0
	depth := uint8(0)

	for depth = uint8(0); depth < maxDigits; depth++ {
		interval = right.Interval(left, depth) - 1
		fmt.Printf("interval(%d) = %d\n", depth, interval)
		if interval >= 1 {
			break
		}
	}
	if debug && depth >= maxDigits {
		panic("max depth reached")
	}

	var step int
	if boundary < interval {
		step = boundary
	} else {
		step = interval
	}
	delta := rand.Intn(step) + 1

	var prefix *Position
	var digit int
	switch s := getStrategy(m, depth); s {
	case boundaryLoStrategy:
		prefix = left.prefix(depth)
		digit = left.DigitAt(depth) + delta
	case boundaryHiStrategy:
		prefix = right.prefix(depth)
		digit = right.DigitAt(depth) - delta
	default:
		// print("go strategy %s", s)
		panic(fmt.Sprintf("unknown strategy %#v", s))
	}

	// fmt.Printf("prefix = %#v\n", prefix)
	// fmt.Printf("digit = %#v\n", digit)
	// fmt.Printf("depth = %#v\n", depth)
	// fmt.Printf("left index = %#v\n", left.IndexAt(depth))
	// fmt.Printf("right index = %#v\n", right.IndexAt(depth))
	if debug && (digit < 0 || digit > int(maxDigitAtDepth(depth))) {
		panic("generated bad digit")
	}

	out := prefix.Add(uint(digit), site)
	if debug && !(left.IsBefore(out) && out.IsBefore(right)) {
		fmt.Printf("left = %#v\n", left)
		fmt.Printf("out = %#v\n", out)
		fmt.Printf("right = %#v\n", right)
		panic("allocated position not in order")
	}

	return out
}

// GoString --
// Implement `fmt.GoStringer` so that the `%#v` placeholder works for `Position`
// values.
func (pos *Position) GoString() string {
	l := make([]string, pos.length)
	for d := uint8(0); d < pos.length; d++ {
		l[d] = fmt.Sprintf("%d", pos.DigitAt(d))
	}
	return fmt.Sprintf("<%s>", strings.Join(l, ", "))
}
