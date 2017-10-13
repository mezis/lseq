package lseq

import "github.com/Workiva/go-datastructures/common"

type atom struct {
	pos  *Position // position identifier
	data string    // the actual text
}

func newAtom(p *Position, d string) *atom {
	out := new(atom)
	out.pos = p
	out.data = d
	return out
}

func (a *atom) Compare(b common.Comparator) int {
	return a.pos.Compare(b.(*atom).pos)
}
