package wal_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("Entry", func() {
	It("should fail when reading from an empty buffer", func() {
		var buffer bytes.Buffer
		Expect(wal.ReadEntry(&buffer, nil, 1024)).Error().To(MatchError(wal.ErrEntryNone))
	})

	It("should fail when reading a zero only data entry", func() {
		// This test simulates the situation where a segment is only partially filled and the rest of the file which
		// was pre-allocated consists of null bytes.
		buffer := bytes.NewBuffer(make([]byte, 1024))
		Expect(wal.ReadEntry(buffer, nil, int64(buffer.Len()))).Error().To(MatchError(wal.ErrEntryNone))
	})
})
