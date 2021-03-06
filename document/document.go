package document

import (
	"github.com/Workiva/go-datastructures/common"
	"github.com/Workiva/go-datastructures/slice/skip"
	"github.com/mezis/lseq/position"
	"github.com/mezis/lseq/uid"
)

// Document is a mutable ordered lists of atoms (e.g lines, characters)
type Document struct {
	uid.Uid
	atoms *skip.SkipList
	alloc *position.Allocator
}

type atom struct {
	pos  *position.Position // position identifier
	data string             // the actual text
}

func newAtom(p *position.Position, d string) *atom {
	out := new(atom)
	out.pos = p
	out.data = d
	return out
}

func (a *atom) Compare(b common.Comparator) int {
	return a.pos.Compare(b.(*atom).pos)
}

// NewDocument returns a new document
//
// Internally, this has two unremovable atoms - "start" and "stop" sentinels
func NewDocument() *Document {
	doc := new(Document)
	doc.Uid = uid.Generate()
	doc.atoms = skip.New(uint8(0))
	doc.atoms.Insert(newAtom(position.SentinelHead, ""))
	doc.atoms.Insert(newAtom(position.SentinelTail, ""))
	doc.alloc = position.NewAllocator()
	return doc
}

// Length returns the current number of atoms in the document.
func (doc *Document) Length() int {
	return int(doc.atoms.Len()) - 2
}

// Data returns all the atom data currently in the document, in order.
func (doc *Document) Data() []string {
	out := make([]string, doc.Length())
	doc.Each(func(k uint, _ *position.Position, data string) {
		out[k] = data
	})
	return out
}

// Insert  adds a new atom with position `pos` and content `data`.
//
// Returns false if `pos` already exists in the document (and in that case, adds
// nothing)
func (doc *Document) Insert(pos *position.Position, data string) bool {
	a := newAtom(pos, data)
	res := doc.atoms.Insert(a)
	return len(res) == 0
}

// Delete removes the atom referenced `pos` from the document.
//
// Returns true iff the position was present.
func (doc *Document) Delete(pos *position.Position) bool {
	a := atom{pos: pos}
	res := doc.atoms.Delete(&a)
	return len(res) == 1
}

// Each iterates through atoms, passing them to the "cb" callback.
// Skips the first and last "sentinel" atoms.
func (doc *Document) Each(cb func(number uint, pos *position.Position, data string)) {
	head := doc.atoms.ByPosition(1)
	n := doc.Length()
	if head == nil {
		return
	}
	iter := doc.atoms.Iter(head)
	for k := 0; k < n; k++ {
		iter.Next()
		a := iter.Value().(*atom)
		// fmt.Println("iter:", k, a.pos, a.data)
		cb(uint(k), a.pos, a.data)
	}
}

// At returns the atom indexed `idx`.
func (doc *Document) At(idx int) (*position.Position, string) {
	if debug && (idx < 0 || idx >= doc.Length()) {
		panic("index out of bounds")
	}

	a := doc.atoms.ByPosition(uint64(idx + 1)).(*atom)
	return a.pos, a.data
}

// Allocate returns positions ordered immediately before the atom at index `idx`.
// The resulting slice is ordered.
func (doc *Document) Allocate(idx int, count int, site uid.Uid) []*position.Position {
	if debug && (idx < 0 || idx >= doc.Length()) {
		panic("index out of bounds")
	}
	if count < 0 {
		panic("cannot allocate a negative number of positions")
	}

	out := make([]*position.Position, count)

	left := doc.atoms.ByPosition(uint64(idx)).(*atom).pos
	right := doc.atoms.ByPosition(uint64(idx + 1)).(*atom).pos

	for k := range out {
		p := new(position.Position)
		doc.alloc.Call(p, left, right, site)
		out[k] = p
		left = p
	}
	return out
}
