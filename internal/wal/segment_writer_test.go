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
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 1*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 2*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 3*(4+3+4))))
	})
})

func BenchmarkSegmentWriter_AppendEntry(b *testing.B) {
	for _, entryLengthEncoding := range wal.EntryLengthEncodings {
		for _, entryChecksumType := range wal.EntryChecksumTypes {
			for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
				data := make([]byte, dataSize*1024)
				segmentWriter, err := wal.NewSegmentWriter(&SegmentWriterFileDiscard{}, wal.Header{
					Magic:               wal.Magic,
					Version:             wal.HeaderVersion,
					EntryLengthEncoding: entryLengthEncoding,
					EntryChecksumType:   entryChecksumType,
					FirstSequenceNumber: 0,
				}, 0, 0, &wal.SyncPolicyNone{})
				if err != nil {
					b.Fatal(err)
				}
				b.Run(fmt.Sprintf("%s %s %d KB", entryLengthEncoding, entryChecksumType, dataSize), func(b *testing.B) {
					for b.Loop() {
						if err := segmentWriter.AppendEntry(data); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}

// SegmentWriterFileDiscard provides a stub for a segment file which discards all data. It allows us to run large scale
// benchmarks without filling up the disk or memory.
type SegmentWriterFileDiscard struct{}

// SegmentWriterFileDiscard implements SegmentWriterFile.
var _ wal.SegmentWriterFile = (*SegmentWriterFileDiscard)(nil)

func (s *SegmentWriterFileDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (s *SegmentWriterFileDiscard) Close() error {
	return nil
}

func (s *SegmentWriterFileDiscard) Sync() error {
	return nil
}

func (s *SegmentWriterFileDiscard) Name() string {
	return "in-memory-discard"
}
