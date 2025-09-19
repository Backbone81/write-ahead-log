package wal_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/write-ahead-log/internal/wal"
)

var _ = Describe("Init", func() {
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = os.MkdirTemp("", "test-init-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should initialize a write-ahead log", func() {
		Expect(wal.IsInitialized(dir)).To(BeFalse())

		Expect(wal.Init(dir)).To(Succeed())

		Expect(wal.IsInitialized(dir)).To(BeTrue())
	})
})
