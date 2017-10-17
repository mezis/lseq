package document

import (
	"fmt"
	"strings"

	"github.com/mezis/lseq/position"
	"github.com/mezis/lseq/uid"
	"github.com/pmezard/go-difflib/difflib"
)

type patchOp bool

const (
	patchOpDelete patchOp = false
	patchOpInsert patchOp = true
)

type patchItem struct {
	op   patchOp
	pos  *position.Position
	data string
}

// type patchId [16]byte

type patch struct {
	// id    patchId // hash of patch items
	items []patchItem
}

func (p *patch) add(op patchOp, pos *position.Position, data string) {
	p.items = append(p.items, patchItem{op, pos, data})
}

func (p *patch) Length() int {
	return len(p.items)
}

func (p *patch) String() string {
	buf := make([]string, len(p.items))
	for k, i := range p.items {
		buf[k] = fmt.Sprintf("%v\n%v%v", i.pos, i.op, i.data)
	}
	return strings.Join(buf, "\n")
}

// NewPatch returns a new `patch` that, when applied, transforms the text of `doc` into the
// argument list of atoms.
func NewPatch(doc *Document, site uid.Uid, data []string) *patch {
	out := new(patch)

	matcher := difflib.NewMatcher(doc.Data(), data)

	for _, op := range matcher.GetOpCodes() {
		switch op.Tag {
		case 'r':
			for i := op.I1; i < op.I2; i++ {
				p, s := doc.At(i)
				out.add(patchOpDelete, p, s)
			}
			pos := doc.Allocate(op.I2, op.J2-op.J1, site)
			for j := op.J1; j < op.J2; j++ {
				out.add(patchOpInsert, pos[j-op.J1], data[j])
			}
		case 'd':
			for i := op.I1; i < op.I2; i++ {
				p, s := doc.At(i)
				out.add(patchOpDelete, p, s)
			}
		case 'i':
			pos := doc.Allocate(op.I2, op.J2-op.J1, site)
			for j := op.J1; j < op.J2; j++ {
				out.add(patchOpInsert, pos[j-op.J1], data[j])
			}
		default: // 'e' (equal) tag, nothing to do
		}
	}
	return out
}

// Apply
// iterates through patch items and applies them all to the argument Document.
func (p *patch) Apply(doc *Document) {
	for _, i := range p.items {
		switch i.op {
		case patchOpInsert:
			doc.Insert(i.pos, i.data)
		case patchOpDelete:
			doc.Delete(i.pos)
		default:
			panic(fmt.Sprintf("unknown patch operation %#v", i.op))
		}
	}
}
