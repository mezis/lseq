package document_test

import (
	. "github.com/mezis/lseq/document"
	"github.com/mezis/lseq/uid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("patch", func() {
	site := uid.Uid(0x00)
	Context("Given an empty document", func() {
		Describe("NewPatch", func() {
			It("build an empty patch for empty documents", func() {
				left := NewDocument()
				right := []string{}
				p := NewPatch(left, site, right)

				Expect(p.Length()).To(Equal(0))
			})

			It("build a patch of length 2 when adding 2 atoms", func() {
				left := NewDocument()
				right := []string{"hello", "world"}
				p := NewPatch(left, site, right)

				Expect(p.Length()).To(Equal(2))
			})
		})
	})
	Context("Given an initial document", func() {
		data := []string{"hello", "beautiful", "world"}
		buildDocument := func() *Document {
			out := NewDocument()
			p := NewPatch(out, site, data)
			p.Apply(out)
			return out
		}

		It("builds the initial document", func() {
			doc := buildDocument()
			Expect(doc.Data()).To(Equal(data))
		})

		Describe("patch.Apply", func() {
			check := func(target []string) {
				doc := buildDocument()
				p := NewPatch(doc, site, target)
				p.Apply(doc)

				Expect(doc.Data()).To(Equal(target))
			}

			It("inserts lines", func() {
				check([]string{"hello", "beautiful", "world", "of", "mine"})
			})

			It("deletes lines", func() {
				check([]string{"hello", "world"})
			})

			It("replaces lines", func() {
				check([]string{"hello", "frabjous", "world"})
			})

		})
	})

})
