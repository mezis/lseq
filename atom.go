package lseq

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
