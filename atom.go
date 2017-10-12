package lseq

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
