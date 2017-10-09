package lseq

import (
	"sort"

	"github.com/juju/errors"
	"github.com/mezis/lseq/uid"
)

type atom struct {
	pos       *Position // position identifier
	data      string    // the actual text
	tombstone bool      // whether the atom was flagged as deleted
}

func newAtom(p *Position, d string) *atom {
	out := new(atom)
	out.pos = p
	out.data = d
	return out
}

// Document models a mutable, ordered set of strings, which can be added or
// deleted, or listed.
type Document interface {
	// responding to events
	Add(pos *Position, data string)
	Delete(pos *Position)

	// Iterate through document, calling "cb" for each (non-tombstoned) atom.
	Each(cb func(number uint, pos *Position, data string))

	// To permit edition: insert an atom with the "data" before atom at "pos".
	// Return "false" as second argument if the input position is not part of
	// the document or is the sentinel (head) position.
	Insert(site uid.Uid, pos *Position, data string) error
}

// documents are mutable ordered lists of atoms
type document struct {
	uid.Uid
	atoms []*atom
}

// NewDocument returns a new document, with two unremovable atoms - "start" and
// "stop" sentinel strings.
func NewDocument() Document {
	headPos := new(Position).Add(0, 0)
	tailPos := new(Position).Add(maxIndexAtDepth(0), 0)
	if headPos == nil || tailPos == nil {
		panic("could not create positions")
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

func (doc *document) Insert(site uid.Uid, pos *Position, data string) error {
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

func (doc *document) Add(pos *Position, data string) {
	return
}

func (doc *document) Delete(pos *Position) {
	return
}

func (doc *document) Each(cb func(number uint, pos *Position, data string)) {
	var number uint // = 0
	for _, a := range doc.atoms {
		if a.tombstone {
			continue
		}
		cb(number, a.pos, a.data)
		number++
	}
}
