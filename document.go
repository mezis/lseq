package lseq

import (
	"sort"

	"github.com/juju/errors"
	"github.com/mezis/lseq/uid"
)

// documents are mutable ordered lists of atoms
type Document struct {
	uid.Uid
	atoms         []*atom     // the ordered sequence of atoms
	allocStrategy StrategyMap // depth -> alloc strategy
}

// NewDocument returns a new document, with two unremovable atoms - "start" and
// "stop" sentinel strings.
func NewDocument() *Document {
	headPos := new(Position).Add(0, 0)
	tailPos := new(Position).Add(maxIndexAtDepth(0), 0)
	if headPos == nil || tailPos == nil {
		panic("could not create positions")
	}

	doc := Document{Uid: uid.New(), atoms: make([]*atom, 0, 2)}
	doc.addAtom(newAtom(headPos, ""))
	doc.addAtom(newAtom(tailPos, ""))
	return &doc
}

// Add the atom in the sorted array
func (doc *Document) addAtom(a *atom) error {
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
	return nil
}

func (doc *Document) Insert(site uid.Uid, pos *Position, data string) error {
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
	newPos := new(Position)

	// extend array and add new atom
	head := doc.atoms[:idx]
	tail := doc.atoms[idx:]
	doc.atoms = make([]*atom, len(doc.atoms)+1)
	copy(doc.atoms[:idx], head)
	copy(doc.atoms[idx+1:], tail)
	doc.atoms[idx] = newAtom(newPos, data)

	return nil
}

func (doc *Document) Each(cb func(number uint, pos *Position, data string)) {
	var number uint // = 0
	for _, a := range doc.atoms {
		if a.tombstone {
			continue
		}
		cb(number, a.pos, a.data)
		number++
	}
}
