package document_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	. "github.com/mezis/lseq/document"
	"github.com/mezis/lseq/uid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("lseq", func() {
	site := uid.Uid(0)

	Describe("NewDocument", func() {
		It("passes", func() {
			x := NewDocument()
			Expect(x).NotTo(Equal(nil))
		})

		It("has length zero", func() {
			x := NewDocument()
			Expect(x.Length()).To(Equal(0))
		})
	})

	Describe("Document.Allocate", func() {
		x := NewDocument()
		It("returns a slice of positions", func() {
			res := x.Allocate(0, 10, site)
			Expect(res).NotTo(Equal(nil))
			Expect(len(res)).To(Equal(10))
		})

		It("returns ordered positions", func() {
		})
	})

	buildDocument := func() *Document {
		data := []string{"foo", "bar", "qux"}
		out := NewDocument()
		pos := out.Allocate(0, len(data), site)
		for k, s := range data {
			out.Insert(pos[k], s)
		}
		return out
	}

	Describe("Document.At", func() {
		It("returns the Nth position and data entry", func() {
			doc := buildDocument()
			p, s := doc.At(1)
			Expect(p).NotTo(Equal(nil))
			Expect(s).To(Equal("bar"))
		})
	})

	Describe("Document.Insert", func() {
		It("increases the atom count", func() {
			doc := buildDocument()
			Expect(doc.Length()).To(Equal(3))
		})

		It("adds the data", func() {
			doc := buildDocument()
			Expect(doc.Data()).To(Equal([]string{"foo", "bar", "qux"}))
		})

		XIt("returns false if the atom already existed", func() {})
	})
	Describe("Document.Delete", func() {
		perform := func() (*Document, interface{}) {
			doc := buildDocument()
			p, _ := doc.At(1)
			return doc, doc.Delete(p)
		}

		It("returns true", func() {
			_, res := perform()
			Expect(res).To(BeTrue())
		})

		It("decreases the line count", func() {
			doc, _ := perform()
			Expect(doc.Length()).To(Equal(2))
		})

		It("removes the data", func() {
			doc, _ := perform()
			Expect(doc.Data()).To(Equal([]string{"foo", "qux"}))
		})

		XIt("returns false when the item didn't exist", func() {})
	})

	Describe("Document.Each", func() {
		XIt("iterates over all items", func() {})
	})
})

func BenchmarkDocumentRandomEdits(b *testing.B) {
	for _, exp := range []uint{10, 11, 12, 13, 14, 15, 16} {
		count := 1 << exp
		doc := NewDocument()
		for k, pos := range doc.Allocate(0, count, 0x00) {
			str := fmt.Sprintf("atom%04d", k)
			doc.Insert(pos, str)
		}
		if doc.Length() != count {
			panic("benchmark doc has wrong length")
		}
		runtime.GC()

		b.Run(fmt.Sprintf("N=%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for k := 0; k < b.N; k++ {
				n := rand.Intn(doc.Length())
				p, _ := doc.At(n)
				str := fmt.Sprintf("edit%05d", k)
				q := doc.Allocate(n, 1, 0x00)
				doc.Delete(p)
				doc.Insert(q[0], str)
			}
		})
	}
}
