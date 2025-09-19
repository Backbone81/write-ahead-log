package encoding_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encoding Suite")
}
