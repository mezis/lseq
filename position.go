package lseq

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/mezis/lseq/uid"
)

// Position - An immutable position in a document
type Position struct {
	// A variable-base number; the first digit is the `rootBits` most
	// significant bits of `value`; the second digit is the next
	// `rootBits+1` bits, etc.
	digits big.Int

	// The number of variable-base digits in `value`
	length uint8

	// Same as `value`, but interleaved with site identifiers (`uid.Uid`).
	// This allows for fully ordered comparison between positions.
	sites big.Int
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

// Precalculated 64-bit mask
var siteMask = func() *big.Int {
	one := big.NewInt(1)
	two64 := new(big.Int).Lsh(one, uid.Bits)
	mask := new(big.Int).Sub(two64, one)
	return mask
}()

// Precalculated masks for digits with variable bases
var digitMask = func() []*big.Int {
	out := make([]*big.Int, maxDigits)
	for d := uint8(0); d < maxDigits; d++ {
		out[d] = big.NewInt(int64(maxDigitAtDepth(d)))
	}
	return out
}()

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// // Change the receiver to be identical to `oth`
// func (pos *Position) clone(oth *Position) {
// 	pos.length = oth.length
// 	pos.digits.Set(&oth.digits)
// 	pos.sites.Set(&oth.sites)
// }

// // Change the position to be length at least `length`, with the `length` most
// // significant digits taken from `oth`.
// //
// // Extra digits and their site identifiers are set to zero.
// func (pos *Position) padTo(oth *Position, length uint8) {
// 	pos.length = oth.length

// 	if pos.length >= length {
// 		pos.digits.Set(&oth.digits)
// 		pos.sites.Set(&oth.sites)
// 		return
// 	}

// 	var shiftDigitsBy, shiftSitesBy uint // = 0
// 	for pos.length < length {
// 		digitBits := uint(bitsAtDepth(pos.length))
// 		shiftDigitsBy += digitBits
// 		shiftSitesBy += digitBits + uid.Bits
// 		pos.length++
// 	}
// 	pos.digits.Lsh(&oth.digits, shiftDigitsBy)
// 	pos.sites.Lsh(&oth.sites, shiftSitesBy)
// }

// // Change the position to have length at most `length`, taking up to `length` of
// // the most significant digits or `oth`.
// func (pos *Position) trimTo(oth *Position, length uint8) {
// 	pos.length = oth.length

// 	if pos.length <= length {
// 		pos.digits.Set(&oth.digits)
// 		pos.sites.Set(&oth.sites)
// 		return
// 	}

// 	var shiftDigitsBy, shiftSitesBy uint // = 0
// 	for pos.length > length {
// 		digitBits := uint(bitsAtDepth(pos.length - 1))
// 		shiftDigitsBy += digitBits
// 		shiftSitesBy += digitBits + uid.Bits
// 		pos.length--
// 	}
// 	pos.digits.Rsh(&oth.digits, shiftDigitsBy)
// 	pos.sites.Rsh(&oth.sites, shiftSitesBy)
// }

// // Set the position to be a prefix of the `length` most significant digits of
// // `oth`, padding with zeroes or removing digits as appropriate.
// func (pos *Position) prefix(oth *Position, length uint8) {
// 	if oth.length < length {
// 		pos.padTo(oth, length)
// 	} else if oth.length > length {
// 		pos.trimTo(oth, length)
// 	} else {
// 		pos.clone(oth)
// 	}
// }

// IsBefore -
// Return true iff "pos" is before "oth" in the partial order defined by Logoot.
//
// In practice, this is the lexicographical order on the list of (identifier,
// site) pairs, implemented as integer comparison after padding.
func (pos *Position) IsBefore(oth *Position) bool {
	// how many bits to left-shift `sites` by, if currently of length `a`,
	// to be length `b` ?
	padBits := func(a uint8, b uint8) uint {
		out := uint(0)
		for l := a; l < b; l++ {
			out += uint(bitsAtDepth(l)) + uid.Bits
		}
		return out
	}

	posPad := padBits(pos.length, oth.length)
	othPad := padBits(oth.length, pos.length)

	// pad both positions' sites, and reset after comparing (to avoids allocations
	// of padded positions)
	if posPad > 0 {
		pos.sites.Lsh(&pos.sites, posPad)
	}
	if othPad > 0 {
		oth.sites.Lsh(&oth.sites, othPad)
	}

	res := pos.sites.Cmp(&oth.sites) < 0

	if posPad > 0 {
		pos.sites.Rsh(&pos.sites, posPad)
	}
	if othPad > 0 {
		oth.sites.Rsh(&oth.sites, othPad)
	}

	return res
}

func (pos *Position) equals(oth *Position) bool {
	return pos.length == oth.length && pos.sites.Cmp(&oth.sites) == 0
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

	out.digits.Lsh(&pos.digits, digitBits)
	out.digits.Or(&out.digits, new(big.Int).SetUint64(uint64(digit)))

	out.sites.Lsh(&pos.sites, digitBits)
	out.sites.Or(&out.sites, new(big.Int).SetUint64(uint64(digit)))
	out.sites.Lsh(&out.sites, uid.Bits)
	out.sites.Or(&out.sites, site.ToBig(new(big.Int)))

	return out
}

// Length --
// Return the number of digits in this position.
func (pos *Position) Length() int {
	return int(pos.length)
}

// put the Nth digit of `pos` into `out`
func (pos *Position) digitAt(out *big.Int, depth uint8) {
	if depth >= pos.length {
		out.SetUint64(uint64(0))
	}
	shiftBy := uint(0)
	for d := pos.length - 1; d > depth; d-- {
		shiftBy += uint(bitsAtDepth(uint8(d)))
	}
	out.Rsh(&pos.digits, shiftBy)
	mask := digitMask[depth]
	out.And(out, mask)
}

// DigitAt -
// Return the value of the `depth`s most significant digit.
func (pos *Position) DigitAt(depth uint8) int {
	if pos.length <= depth {
		return 0
	}

	var val big.Int
	pos.digitAt(&val, depth)
	return int(val.Int64())
}

func (pos *Position) siteAt(out *big.Int, depth uint8) {
	if depth >= pos.length {
		out.SetUint64(uint64(0))
	}
	shiftBy := uint(0)
	for d := pos.length - 1; d > depth; d-- {
		shiftBy += uint(bitsAtDepth(uint8(d))) + uid.Bits
	}

	out.Rsh(&pos.sites, shiftBy)
	out.And(out, siteMask)
}

// SiteAt -
// Return the value of the site identifier for the `depth`'s most significant
// digit.
func (pos *Position) SiteAt(depth uint8) uid.Uid {
	if pos.length <= depth {
		return 0
	}

	var val big.Int
	pos.siteAt(&val, depth)
	return uid.New(&val)
}

// Interval -
// Return the distance, in number of free identifiers, between "pos" and "oth"
func (pos *Position) Interval(oth *Position) int {
	if debug && pos.length != oth.length {
		// TODO: this could be supported with padding. necessary?
		panic("positions have different lengths")
	}
	delta := new(big.Int).Sub(&pos.digits, &oth.digits)
	if debug && !delta.IsInt64() {
		panic(fmt.Sprintf("interval from %#v to %#v is out of bounds", oth, pos))
	}
	interval := int(delta.Int64()) - 1

	return max(0, interval)
}

// String --
// Implement `fmt.Stringer` so that the `%v` placeholder works for `Position`
// values.
func (pos *Position) String() string {
	l := make([]string, pos.length)
	for d := uint8(0); d < pos.length; d++ {
		l[d] = fmt.Sprintf("%d %#v", pos.DigitAt(d), pos.SiteAt(d))
	}
	return fmt.Sprintf("<%s>", strings.Join(l, ", "))
}

// GoString --
// Implement `fmt.GoStringer` so that the `%v` placeholder works for `Position`
// values.
func (pos *Position) GoString() string {
	return pos.String()
}
