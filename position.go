package lseq

import (
	"math/big"

	"github.com/juju/errors"
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

// Number of bits used for the indexes at the root of the tree (depth zero)
const rootBits uint = 5

// Maximum tree depth (ie. position length); results in 31-bit indexes at the
// deepest level.
const maxLength = 26

// How many free identifiers to leave before or after the
const boundary = 10

// The maximum index value at a given tree depth (ie. position length).
// Note that is is also usable as a bitmask.
func maxIndex(depth uint8) uint {
	return 1<<(uint(depth)+rootBits) - 1
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
	return pos
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
	return pos
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
func (pos *Position) Add(index uint, site uid.Uid) (*Position, error) {
	if pos.length >= maxLength {
		return nil, errors.New("max position length reached")
	}

	indexBits := rootBits + uint(pos.length)
	if index < 0 || index >= 1<<indexBits {
		return nil, errors.New("bad index value")
	}

	out := new(Position)
	out.length = pos.length + 1
	out.value.Lsh(&pos.value, indexBits)
	out.value.Or(&out.value, new(big.Int).SetUint64(uint64(index)))
	out.value.Lsh(&out.value, uid.Bits)
	out.value.Or(&out.value, site.ToBig())

	return out, nil
}

// Return the last (deepest) index value.
func (pos *Position) lastIndex() uint64 {
	mask := new(big.Int).SetUint64(uint64(maxIndex(pos.length - 1)))
	val := new(big.Int).Rsh(&pos.value, uid.Bits)
	val.And(val, mask)
	return val.Uint64()
}

// Interval -
// Return the distance, in number of free identifiers, between "pos" and "oth".
// It is expected that "pos" ≺ "oth".
func (pos *Position) Interval(oth *Position, depth uint8) int64 {
	iPos := pos.prefix(depth).lastIndex()
	iOth := oth.prefix(depth).lastIndex()
	return int64(iOth) - int64(iPos)
}
