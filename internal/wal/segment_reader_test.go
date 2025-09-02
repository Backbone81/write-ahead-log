package wal_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("SegmentReader", func() {
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = os.MkdirTemp("", "test-segment-writer-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should read an empty segment file", func() {
		writer, err := wal.CreateSegment(dir, 7, 0, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.OpenSegment(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
	})

	It("should read a full segment file", func() {
		writer, err := wal.CreateSegment(dir, 7, 0, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("bar"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("baz"))).To(Succeed())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.OpenSegment(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Err()).To(Succeed())
		Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
			SequenceNumber: 7,
			Data:           []byte("foo"),
		}))

		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Err()).To(Succeed())
		Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
			SequenceNumber: 8,
			Data:           []byte("bar"),
		}))

		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Err()).To(Succeed())
		Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
			SequenceNumber: 9,
			Data:           []byte("baz"),
		}))

		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
	})

	It("should read a pre-allocated segment file", func() {
		writer, err := wal.CreateSegment(dir, 7, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.OpenSegment(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).ToNot(MatchError(io.EOF))
		Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))
	})

	It("should correctly report sequence numbers", func() {
		writer, err := wal.CreateSegment(dir, 7, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.OpenSegment(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.NextSequenceNumber()).To(Equal(uint64(7)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.NextSequenceNumber()).To(Equal(uint64(8)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.NextSequenceNumber()).To(Equal(uint64(9)))
		Expect(reader.Next()).To(BeTrue())
	})

	It("should correctly report offsets", func() {
		writer, err := wal.CreateSegment(dir, 7, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.AppendEntry([]byte("foo"))).To(Succeed())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.OpenSegment(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 1*(4+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 2*(4+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 3*(4+3+4))))
	})
})

func BenchmarkSegmentReader_Next(b *testing.B) {
	for _, entryLengthEncoding := range wal.EntryLengthEncodings {
		for _, entryChecksumType := range wal.EntryChecksumTypes {
			for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
				data := make([]byte, dataSize*1024)
				recorder := SegmentWriterFileRecorder{}
				segmentWriter, err := wal.NewSegmentWriter(&recorder, wal.Header{
					Magic:               wal.Magic,
					Version:             1,
					EntryLengthEncoding: entryLengthEncoding,
					EntryChecksumType:   entryChecksumType,
					FirstSequenceNumber: 0,
				}, 0, 0, &wal.SyncPolicyNone{})
				if err != nil {
					b.Fatal(err)
				}

				if err := segmentWriter.AppendEntry(data); err != nil {
					b.Fatal(err)
				}

				readerLoop := SegmentReaderFileLoop{
					Data: recorder.Bytes(),
				}
				segmentReader, err := wal.NewSegmentReader(&readerLoop, segmentWriter.Header(), math.MaxInt64, 0, 0)
				if err != nil {
					b.Fatal(err)
				}
				b.Run(fmt.Sprintf("%s %s %d KB", entryLengthEncoding, entryChecksumType, dataSize), func(b *testing.B) {
					for b.Loop() {
						if !segmentReader.Next() {
							b.Fatal("segment reader could not make progress")
						}
					}
				})
			}
		}
	}
}

// SegmentReaderFileLoop provides a stub for the segment file which returns the same data over and over again in an
// endless loop. It allows us to run large scale benchmarks without having to provide an actual big file on disk or
// memory.
type SegmentReaderFileLoop struct {
	Data   []byte
	Offset int
}

// SegmentReaderFileLoop implements SegmentReaderFile.
var _ wal.SegmentReaderFile = (*SegmentReaderFileLoop)(nil)

func (s *SegmentReaderFileLoop) Read(p []byte) (int, error) {
	copyBytes := min(len(p), len(s.Data)-s.Offset)
	copy(p, s.Data[s.Offset:s.Offset+copyBytes])
	s.Offset += copyBytes
	if s.Offset >= len(s.Data) {
		s.Offset = 0
	}
	return copyBytes, nil
}

func (s *SegmentReaderFileLoop) Close() error {
	return nil
}

func (s *SegmentReaderFileLoop) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (s *SegmentReaderFileLoop) Name() string {
	return "in-memory-loop"
}

// SegmentWriterFileRecorder provides a stub for a segment file which records what is written to it in memory. It allows
// us to use a SegmentWriter to prepare a buffer which can then be used by SegmentReaderFileLoop to serve read requests.
type SegmentWriterFileRecorder struct {
	bytes.Buffer
}

// SegmentWriterFileRecorder implements SegmentWriterFile.
var _ wal.SegmentWriterFile = (*SegmentWriterFileRecorder)(nil)

func (s *SegmentWriterFileRecorder) Close() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Sync() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Name() string {
	return "in-memory-recorder"
}
