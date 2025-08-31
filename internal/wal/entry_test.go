package wal_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("Entry", func() {
	It("should write an entry", func() {
		var buffer bytes.Buffer
		Expect(wal.WriteEntry(&buffer, []byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(buffer.Len()).To(Equal(15))
	})

	It("should read an entry", func() {
		var buffer bytes.Buffer
		Expect(wal.WriteEntry(&buffer, []byte("foo"))).Error().ToNot(HaveOccurred())

		_, data, err := wal.ReadEntry(&buffer, nil, int64(buffer.Len()))
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(Equal([]byte("foo")))
	})

	It("should fail when reading from an empty buffer", func() {
		var buffer bytes.Buffer
		Expect(wal.ReadEntry(&buffer, nil, 1024)).Error().To(MatchError(wal.ErrEntryNone))
	})

	It("should fail when reading from a partial buffer", func() {
		var buffer bytes.Buffer
		Expect(wal.WriteEntry(&buffer, []byte("foo"))).Error().ToNot(HaveOccurred())
		buffer.Truncate(buffer.Len() - 1)
		Expect(wal.ReadEntry(&buffer, nil, int64(buffer.Len()))).Error().To(MatchError(wal.ErrEntryNone))
	})

	It("should fail when the checksum does not match", func() {
		var buffer bytes.Buffer
		Expect(wal.WriteEntry(&buffer, []byte("foo"))).Error().ToNot(HaveOccurred())
		buffer.Bytes()[buffer.Len()-1] = buffer.Bytes()[buffer.Len()-1] + 1
		Expect(wal.ReadEntry(&buffer, nil, int64(buffer.Len()))).Error().To(MatchError(wal.ErrEntryNone))
	})

	PIt("should fail when reading a zero only data entry", func() {
		// This test simulates the situation where a segment is only partially filled and the rest of the file which
		// was pre-allocated consists of null bytes.
		buffer := bytes.NewBuffer(make([]byte, 1024))
		Expect(wal.ReadEntry(buffer, nil, int64(buffer.Len()))).Error().To(MatchError(wal.ErrEntryNone))
	})
})

func BenchmarkEntry_Write(b *testing.B) {
	for _, i := range []int{0, 1, 2, 4, 8} {
		b.Run(fmt.Sprintf("%d KB data", i), func(b *testing.B) {
			// We make sure that our buffer is big enough for holding b.N write operations. Otherwise, our benchmark
			// also measures memory allocations when increasing the buffer size during writes.
			buffer := bytes.NewBuffer(make([]byte, 0, b.N*(8+i*1024+4)))
			data := make([]byte, i*1024)
			b.ResetTimer()

			for range b.N {
				if _, err := wal.WriteEntry(buffer, data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEntry_Read(b *testing.B) {
	for _, i := range []int{0, 1, 2, 4, 8} {
		b.Run(fmt.Sprintf("%d KB data", i), func(b *testing.B) {
			// We make sure that our buffer is big enough for holding b.N write operations. Otherwise, our benchmark
			// also measures memory allocations when increasing the buffer size during writes.
			buffer := bytes.NewBuffer(make([]byte, 0, b.N*(8+i*1024+4)))
			data := make([]byte, i*1024)
			for range b.N {
				if _, err := wal.WriteEntry(buffer, data); err != nil {
					b.Fatal(err)
				}
			}
			b.ResetTimer()

			for range b.N {
				if _, _, err := wal.ReadEntry(buffer, data, int64(buffer.Len())); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
