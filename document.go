package lseq

import (
	"sort"

	"github.com/mezis/lseq/uid"
)

// Document - a mutable ordered lists of atoms (e.g lines, characters)
type Document struct {
	uid.Uid
	atoms []*atom // the sequence of atoms, always kept sorted
	alloc *Allocator
}

// NewDocument returns a new document, with two unremovable atoms - "start" and
// "stop" sentinel strings.
func NewDocument() *Document {
	headPos := new(Position).Append(0, 0)
	tailPos := new(Position).Append(maxDigitAtDepth(0), 0)
	if headPos == nil || tailPos == nil {
		panic("could not create positions")
	}

	doc := Document{Uid: uid.Generate(), atoms: make([]*atom, 0, 2)}
	doc.addAtom(newAtom(headPos, ""))
	doc.addAtom(newAtom(tailPos, ""))
	doc.alloc = NewAllocator()
	return &doc
}

func (doc *Document) Length() int {
	return len(doc.atoms) - 2
}

func (doc *Document) Data() []string {
	out := make([]string, len(doc.atoms)-2)
	doc.Each(func(k uint, _ *Position, data string) {
		out[k] = data
	})
	return out
}

// Add the atom in the sorted array
func (doc *Document) addAtom(a *atom) {
	// find where to insert atom
	idx := sort.Search(len(doc.atoms), func(k int) bool {
		return a.pos.IsBefore(doc.atoms[k].pos)
	})
	head := doc.atoms[:idx]
	tail := doc.atoms[idx:]
	doc.atoms = make([]*atom, len(doc.atoms)+1)
	copy(doc.atoms[:idx], head)
	copy(doc.atoms[idx+1:], tail)
	doc.atoms[idx] = a
}

// Insert
// add a new atom with position `pos` and content `data`.
// TODO: Returns false if `pos` already exists in the document (and in that case, adds
// nothing)
func (doc *Document) Insert(pos *Position, data string) bool {
	// locate the position just after in the atom array
	idx := sort.Search(len(doc.atoms), func(k int) bool {
		return pos.IsBefore(doc.atoms[k].pos)
	})

	if idx == 0 {
		panic("specified position before head sentinel")
	}
	if idx == len(doc.atoms) {
		panic("specified position after tail sentinel")
	}

	// extend array and add new atom
	// TODO: preprovision more capcity to avoid allocations.
	head := doc.atoms[:idx]
	tail := doc.atoms[idx:]
	doc.atoms = make([]*atom, len(doc.atoms)+1)
	copy(doc.atoms[:idx], head)
	copy(doc.atoms[idx+1:], tail)
	doc.atoms[idx] = newAtom(pos, data)

	return true
}

// Delete
// removes the atom referenced `pos` from the document.
func (doc *Document) Delete(pos *Position) bool {
	idx := sort.Search(len(doc.atoms), func(k int) bool {
		// return a.pos.IsBefore(doc.atoms[k].pos)
		return pos.equals(doc.atoms[k].pos)
	})

	if idx == 0 || idx == len(doc.atoms)-1 {
		panic("cannot remove sentinel atoms")
	}
	if idx == len(doc.atoms) {
		// not found
		return false
	}

	copy(doc.atoms[idx:len(doc.atoms)-1], doc.atoms[idx+1:])
	doc.atoms = doc.atoms[:len(doc.atoms)-1]
	return true
}

// Each --
// Iterate through atoms, passing them to the "cb" callback.
// Skips the first and last "sentinel" atoms.
func (doc *Document) Each(cb func(number uint, pos *Position, data string)) {
	for k, a := range doc.atoms[1 : len(doc.atoms)-1] {
		cb(uint(k), a.pos, a.data)
	}
}

// At
// returns the atom indexed `idx`.
func (doc *Document) At(idx int) (*Position, string) {
	if debug && (idx < 0 || idx >= len(doc.atoms)) {
		panic("index out of bounds")
	}

	a := doc.atoms[idx+1]
	return a.pos, a.data
}

// Allocate
// returns positions ordered immediately after the atom at index `idx`.
// The return slice is ordered.
func (doc *Document) Allocate(idx int, count int, site uid.Uid) []*Position {
	if debug && (idx < 0 || idx >= len(doc.atoms)) {
		panic("index out of bounds")
	}
	if count < 0 {
		panic("cannot allocate a negative number of positions")
	}

	out := make([]*Position, count)

	left := doc.atoms[idx].pos
	right := doc.atoms[idx+1].pos

	for k := range out {
		p := new(Position)
		doc.alloc.Call(p, left, right, site)
		out[k] = p
		left = p
	}
	return out
}
