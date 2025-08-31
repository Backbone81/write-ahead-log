package wal_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("SegmentWriter", func() {
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = os.MkdirTemp("", "test-segment-writer-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should create a new segment file", func() {
		entriesBefore, err := os.ReadDir(dir)
		Expect(err).ToNot(HaveOccurred())

		writer, err := wal.CreateSegment(dir, 0, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		entriesAfter, err := os.ReadDir(dir)
		Expect(err).ToNot(HaveOccurred())
		Expect(entriesAfter).To(HaveLen(len(entriesBefore) + 1))
	})

	It("should write to the segment file", func() {
		writer, err := wal.CreateSegment(dir, 0, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
	})

	It("should correctly report sequence numbers", func() {
		writer, err := wal.CreateSegment(dir, 7, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.NextSequenceNumber()).To(Equal(uint64(7)))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(8)))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(9)))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(10)))
	})

	It("should correctly report offsets", func() {
		writer, err := wal.CreateSegment(dir, 0, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize)))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 1*(8+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 2*(8+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 3*(8+3+4))))
	})
})

func BenchmarkSegmentWriter_AppendEntry_1000x(b *testing.B) {
	for _, i := range []int{0, 1, 2, 4, 8} {
		b.Run(fmt.Sprintf("%d KB data", i), func(b *testing.B) {
			for range b.N {
				writeSegmentWith1000Entries(b, i*1024)
			}
		})
	}
}

func writeSegmentWith1000Entries(b *testing.B, dataSize int) {
	b.Helper()
	b.StopTimer()

	writer, err := wal.CreateSegment(b.TempDir(), 0, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
	Expect(err).ToNot(HaveOccurred())
	defer func() {
		Expect(writer.Close()).To(Succeed())
	}()

	data := make([]byte, dataSize)

	b.StartTimer()
	for range 1000 {
		if err := writer.AppendEntry(data); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
