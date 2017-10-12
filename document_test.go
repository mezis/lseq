package lseq_test

import (
	"github.com/mezis/lseq"
	"github.com/mezis/lseq/uid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("lseq", func() {
	site := uid.Uid(0)

	Describe("NewDocument", func() {
		It("passes", func() {
			x := lseq.NewDocument()
			Expect(x).NotTo(Equal(nil))
		})
	})

	Describe("Document.Allocate", func() {
		x := lseq.NewDocument()
		It("returns a slice of positions", func() {
			res := x.Allocate(0, 10, site)
			Expect(res).NotTo(Equal(nil))
			Expect(len(res)).To(Equal(10))
		})

		It("returns ordered positions", func() {
		})
	})

	buildDocument := func() *lseq.Document {
		data := []string{"foo", "bar", "qux"}
		out := lseq.NewDocument()
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
		It("increases the line count", func() {
			doc := buildDocument()
			Expect(doc.Length()).To(Equal(3))
		})
		It("adds the data", func() {
			doc := buildDocument()
			Expect(doc.Data()).To(Equal([]string{"foo", "bar", "qux"}))
		})
	})
	Describe("Document.Delete", func() {
		perform := func() (*lseq.Document, interface{}) {
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
	})

	Describe("Document.Each", func() {})
})
