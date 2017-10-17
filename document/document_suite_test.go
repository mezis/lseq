package document_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDocument(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Document Suite")
}
