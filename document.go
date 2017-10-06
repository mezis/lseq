package lseq

import (
	"github.com/juju/errors"
	"github.com/mezis/lseq/uid"
	"math/big"
	"sort"
)

// number of bits used for the indexes at the root of the tree (depth zero)
const rootBits uint = 5

// maximum tree depth (ie. position length); results in 32-bit indexes at the
// deepest level.
const maxLength = 27

// how many free identifiers to leave before or after the
const boundary = 10

// the maximum index value at a given tree depth (ie. position length)
func maxIndex(depth uint) uint {
	return 1<<(depth+rootBits) - 1
}

// an immutable position in a document
type position struct {
	length uint8
	// this "value" is interpreted as a list of 2 * "length" integers.
	// integer 2k has (rootBits+k) bits and is an index;
	// integer 2k+1 has 64 bits and is a "site identifier" (aka. author
	// identifier).
	// each paid forms an "identifier".
	value *big.Int
}

func newPosition() *position {
	return &position{0, new(big.Int)}
}

func (pos *position) clone() *position {
	out := new(position)
	out.length = pos.length
	out.value = new(big.Int).Set(pos.value)
	return out
}

// Return a position of length at least "length", padded with zeros on the right
// as necessary.
// The original position is returned if its length is already at least "length".
func (pos *position) padTo(length uint8) *position {
	if pos.length >= length {
		return pos
	}

	out := pos.clone()

	var shiftBy uint // = 0
	for out.length < length {
		shiftBy += rootBits + uint(length) + uid.Bits
		out.length++
	}
	out.value.Lsh(out.value, shiftBy)
	return out
}

// Return a position of length at most "length", removing trialing identifiers
// as necessary.
// The original position is returned if already "length" or shorter.
func (pos *position) prefix(length uint8) *position {
	if pos.length <= length {
		return pos
	}

	out := pos.clone()

	var shiftBy uint // = 0
	for out.length > length {
		shiftBy += rootBits + uint(out.length-1) + uid.Bits
		out.length--
	}
	out.value.Rsh(out.value, shiftBy)
	return out

}

// ordering of positions = lexicographical order on the list of (identifier,
// site) pairs.
// implemented as integer comparison after padding.
func (pos *position) isBefore(oth *position) bool {
	pos = pos.padTo(oth.length)
	oth = oth.padTo(pos.length)
	return pos.value.Cmp(oth.value) < 0
}

func (pos *position) equals(oth *position) bool {
	return pos.value.Cmp(oth.value) == 0
}

// add an identifier to an existing position and return the new position
func (pos *position) add(index uint, site uid.Uid) (*position, error) {
	if pos.length >= maxLength {
		return nil, errors.New("max position length reached")
	}

	indexBits := rootBits + uint(pos.length)
	if index < 0 || index >= 1<<indexBits {
		return nil, errors.New("bad index value")
	}

	out := new(position)
	out.length = pos.length + 1
	out.value = new(big.Int)
	out.value.Lsh(pos.value, indexBits)
	out.value.Or(out.value, new(big.Int).SetUint64(uint64(index)))
	out.value.Lsh(out.value, uid.Bits)
	out.value.Or(out.value, site.ToBig())

	return out, nil
}

type atom struct {
	pos       *position // position identifier
	data      string    // the actual text
	tombstone bool      // whether the atom was flagged as deleted
}

func newAtom(p *position, d string) *atom {
	out := new(atom)
	out.pos = p
	out.data = d
	return out
}

// Document models a mutable, ordered set of strings, which can be added or
// deleted, or listed.
type Document interface {
	// responding to events
	Add(pos *position, data string)
	Delete(pos *position)

	// Iterate through document, calling "cb" for each (non-tombstoned) atom.
	Each(cb func(number uint, pos *position, data string))

	// To permit edition: insert an atom with the "data" before atom at "pos".
	// Return "false" as second argument if the input position is not part of
	// the document or is the sentinel (head) position.
	Insert(site uid.Uid, pos *position, data string) error
}

// documents are mutable ordered lists of atoms
type document struct {
	uid.Uid
	atoms []*atom
}

// NewDocument returns a new document, with two unremovable atoms - "start" and
// "stop" sentinel strings.
func NewDocument() Document {
	headPos, err := newPosition().add(0, 0)
	if err != nil {
		panic(err.Error())
	}

	tailPos, err := newPosition().add(maxIndex(0), 0)
	if err != nil {
		panic(err.Error())
	}

	doc := document{uid.New(), make([]*atom, 0, 2)}
	doc.addAtom(newAtom(headPos, ""))
	doc.addAtom(newAtom(tailPos, ""))
	return &doc
}

// Add the atom in the sorted array
func (doc *document) addAtom(a *atom) error {
	// find where to insert atom
	idx := sort.Search(len(doc.atoms), func(k int) bool {
		return a.pos.isBefore(doc.atoms[k].pos)
	})
	head := doc.atoms[:idx]
	tail := doc.atoms[idx:]
	doc.atoms = make([]*atom, len(doc.atoms)+1)
	copy(doc.atoms[:idx], head)
	copy(doc.atoms[idx+1:], tail)
	doc.atoms[idx] = a
	return nil
}

func (doc *document) Insert(site uid.Uid, pos *position, data string) error {
	// locate the position in the atom array
	idx := sort.Search(len(doc.atoms), func(k int) bool {
		return pos.equals(doc.atoms[k].pos)
	})

	if idx == 0 || idx == len(doc.atoms) {
		return errors.New("specified position not found")
	}

	// TODO: generate position!
	// left := doc.atoms[idx-1]
	// right := doc.atoms[idx]
	newPos := new(position)

	// extend array and add new atom
	head := doc.atoms[:idx]
	tail := doc.atoms[idx:]
	doc.atoms = make([]*atom, len(doc.atoms)+1)
	copy(doc.atoms[:idx], head)
	copy(doc.atoms[idx+1:], tail)
	doc.atoms[idx] = newAtom(newPos, data)

	return nil
}

func (doc *document) Add(pos *position, data string) {
	return
}

func (doc *document) Delete(pos *position) {
	return
}

func (doc *document) Each(cb func(number uint, pos *position, data string)) {
	var number uint // = 0
	for _, a := range doc.atoms {
		if a.tombstone {
			continue
		}
		cb(number, a.pos, a.data)
		number++
	}
}
