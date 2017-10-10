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

// Return a position of length at least "length", padded with zeros on the right
// as necessary.
// The site identifier of extra digits is set to zero.
//
// Return the receiver if its length is already "length" or more.
func (pos *Position) padTo(length uint8) *Position {
	if pos.length >= length {
		return pos
	}

	out := new(Position)
	out.length = pos.length

	var shiftDigitsBy, shiftSitesBy uint // = 0
	for out.length < length {
		digitBits := uint(bitsAtDepth(out.length))
		shiftDigitsBy += digitBits
		shiftSitesBy += digitBits + uid.Bits
		out.length++
	}
	out.digits.Lsh(&pos.digits, shiftDigitsBy)
	out.sites.Lsh(&pos.sites, shiftSitesBy)
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

	var shiftDigitsBy, shiftSitesBy uint // = 0
	for out.length > length {
		shiftDigitsBy += rootBits + uint(out.length-1)
		shiftSitesBy += rootBits + uint(out.length-1) + uid.Bits
		out.length--
	}
	out.digits.Rsh(&pos.digits, shiftDigitsBy)
	out.sites.Rsh(&pos.sites, shiftSitesBy)
	return out
}

// Return a position of length exactly "length", padding or trimming as
// appropriate.
//
// Return the receiver if already the right length.
func (pos *Position) prefix(length uint8) *Position {
	return pos.trimTo(length).padTo(length)
}

// Return a copy of the position
func (pos *Position) clone() *Position {
	out := new(Position)
	out.length = pos.length
	out.digits.Set(&pos.digits)
	out.sites.Set(&pos.sites)
	return out
}

// IsBefore -
// Return true iff "pos" is before "oth" in the partial order defined by Logoot.
//
// In practice, this is the lexicographical order on the list of (identifier,
// site) pairs, implemented as integer comparison after padding.
func (pos *Position) IsBefore(oth *Position) bool {
	pos = pos.padTo(oth.length)
	oth = oth.padTo(pos.length)
	return pos.sites.Cmp(&oth.sites) < 0
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

	val := new(big.Int).Rsh(&pos.digits, shiftBy)
	val.And(val, digitMask[depth])
	return int(val.Int64())
}

// SiteAt -
// Return the value of the site identifier for the `depth`'s most significant
// digit.
func (pos *Position) SiteAt(depth uint8) uid.Uid {
	if pos.length <= depth {
		return 0
	}

	shiftBy := uint(0)
	for d := pos.length - 1; d > depth; d-- {
		shiftBy += uint(bitsAtDepth(uint8(d))) + uid.Bits
	}

	val := new(big.Int).Rsh(&pos.sites, shiftBy)

	val.And(val, siteMask)
	return uid.New(val)
}

// Interval -
// Return the distance, in number of free identifiers, between "pos" and "oth"
func (pos *Position) Interval(oth *Position) int {
	if debug && pos.length != oth.length {
		panic("positions have different lengths")
	}
	delta := new(big.Int).Sub(&pos.digits, &oth.digits)
	if debug && !delta.IsInt64() {
		panic(fmt.Sprintf("interval from %#v to %#v is out of bounds", oth, pos))
	}
	interval := int(delta.Int64()) - 1

	return max(0, interval)
}

// Iterate through the digits of the receiver, `lt`, and `rt` positions,
// and the site identifiers of the `lt` and `rt` positions.
func (pos *Position) walk(lt *Position, rt *Position, cb func(depth uint8, digit uint64, dlt uint64, drt uint64, slt uid.Uid, srt uid.Uid)) {
	if debug && (pos.length != lt.length || pos.length != rt.length) {
		panic("positions have different lengths")
	}

	var digit, dlt, drt, slt, srt big.Int

	digitsMd := new(big.Int).Set(&pos.digits)
	sitesLt := new(big.Int).Set(&lt.sites)
	sitesRt := new(big.Int).Set(&rt.sites)

	// fmt.Printf("digitMask = (len %d) %#v\n", len(digitMask), digitMask)
	for d := int(pos.length - 1); d >= 0; d-- {
		slt.And(sitesLt, siteMask)
		srt.And(sitesRt, siteMask)

		sitesLt.Rsh(sitesLt, uid.Bits)
		sitesRt.Rsh(sitesRt, uid.Bits)

		// fmt.Println(d)
		// fmt.Println(digitMask[d])
		digit.And(digitsMd, digitMask[d])
		dlt.And(sitesLt, digitMask[d])
		drt.And(sitesRt, digitMask[d])

		digitsMd.Rsh(digitsMd, uint(bitsAtDepth(uint8(d))))
		sitesLt.Rsh(sitesLt, uint(bitsAtDepth(uint8(d))))
		sitesRt.Rsh(sitesRt, uint(bitsAtDepth(uint8(d))))

		cb(uint8(d), digit.Uint64(), dlt.Uint64(), drt.Uint64(), uid.Uid(slt.Uint64()), uid.Uid(srt.Uint64()))
	}
}

// Allocate -
// Implementation of the core LSEQ algorithm. Return a new position between the
// "left" and "right" ones.
func Allocate(left *Position, right *Position, m StrategyMap, site uid.Uid) *Position {
	// fmt.Printf("Allocate(%#v, %#v)\n", left, right)
	if debug && !left.IsBefore(right) {
		panic(fmt.Sprint("arguments not in order ", left, right))
	}

	// find a depth and prefixes with a sufficient interval
	// fmt.Printf("** finding prefixes\n")
	var ltPrefix, rtPrefix *Position
	var interval int
	var depth uint8
	for depth = 1; depth < maxDigits; depth++ {
		// fmt.Printf("*** depth %d\n", depth)
		ltPrefix = left.prefix(depth)
		rtPrefix = right.prefix(depth)
		interval = rtPrefix.Interval(ltPrefix)
		// fmt.Printf("  left  = %#v\n", ltPrefix)
		// fmt.Printf("  right = %#v\n", rtPrefix)
		// fmt.Printf("  interval(%d) = %d\n", depth, interval)
		if interval >= 1 {
			break
		}
	}
	if debug && depth >= maxDigits {
		panic("max depth reached")
	}

	// calcultate digits for the new position
	// fmt.Println("** calculate digits")
	offset := rand.Intn(min(boundary, interval)) + 1

	var out *Position
	bigOffset := big.NewInt(int64(offset))
	switch s := getStrategy(m, depth); s {
	case boundaryLoStrategy:
		out = ltPrefix.clone()
		out.digits.Add(&out.digits, bigOffset)
	case boundaryHiStrategy:
		out = rtPrefix.clone()
		out.digits.Sub(&out.digits, bigOffset)
	default:
		panic(fmt.Sprintf("unknown strategy %#v", s))
	}

	// merge site identifiers
	// fmt.Println("** interleave new indentifiers")
	digits := make([]uint64, depth)
	sites := make([]uid.Uid, depth)
	out.walk(ltPrefix, rtPrefix, func(depth uint8, digit uint64, dlt uint64, drt uint64, slt uid.Uid, srt uid.Uid) {
		digits[depth] = digit
		if digit == dlt {
			sites[depth] = slt
		} else if digit == drt {
			sites[depth] = srt
		} else {
			sites[depth] = site
		}
	})

	// fmt.Println("digits =", digits)
	// fmt.Println("sites =", sites)

	out.sites.SetInt64(0)
	var d, s big.Int
	for k := uint8(0); k < out.length; k++ {
		d.SetUint64(digits[k])
		sites[k].ToBig(&s)

		out.sites.Lsh(&out.sites, uint(bitsAtDepth(k)))
		out.sites.Or(&out.sites, &d)
		out.sites.Lsh(&out.sites, uid.Bits)
		out.sites.Or(&out.sites, &s)
	}

	// check and return
	if debug && !(left.IsBefore(out) && out.IsBefore(right)) {
		fmt.Printf("left = %#v\n", left)
		fmt.Printf("out = %#v\n", out)
		fmt.Printf("right = %#v\n", right)
		panic("allocated position not in order")
	}

	return out
}

// String --
// Implement `fmt.Stringer` so that the `%#v` placeholder works for `Position`
// values.
func (pos *Position) String() string {
	l := make([]string, pos.length)
	for d := uint8(0); d < pos.length; d++ {
		l[d] = fmt.Sprintf("%d %#v", pos.DigitAt(d), pos.SiteAt(d))
	}
	return fmt.Sprintf("<%s>", strings.Join(l, ", "))
}

func (pos *Position) GoString() string {
	return pos.String()
}
