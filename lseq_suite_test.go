package lseq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLseq(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lseq Suite")
}
