package segment_test

import (
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/write-ahead-log/internal/encoding"
	"github.com/backbone81/write-ahead-log/internal/segment"
	"github.com/backbone81/write-ahead-log/internal/utils"
)

var _ = Describe("SegmentReader", func() {
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		for _, entryChecksumType := range encoding.EntryChecksumTypes {
			Context(fmt.Sprintf("With length encoding %s and entry checksum %s", entryLengthEncoding, entryChecksumType), func() {
				var dir string

				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "test-segment-reader-*")
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(dir)).To(Succeed())
				})

				It("should read an empty segment file", func() {
					writer, err := segment.CreateSegment(dir, 0, segment.CreateSegmentConfig{
						PreAllocationSize:   0,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(writer.Close()).To(Succeed())

					reader, err := segment.OpenSegment(dir, 0)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(reader.Close()).To(Succeed())
					}()

					Expect(reader.Next()).To(BeFalse())
					Expect(reader.Err()).To(MatchError(io.EOF))
				})

				It("should read a full segment file", func() {
					writer, err := segment.CreateSegment(dir, 0, segment.CreateSegmentConfig{
						PreAllocationSize:   0,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
					Expect(writer.AppendEntry([]byte("bar"))).Error().ToNot(HaveOccurred())
					Expect(writer.AppendEntry([]byte("baz"))).Error().ToNot(HaveOccurred())
					Expect(writer.Close()).To(Succeed())

					reader, err := segment.OpenSegment(dir, 0)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(reader.Close()).To(Succeed())
					}()

					Expect(reader.Next()).To(BeTrue())
					Expect(reader.Err()).To(Succeed())
					Expect(reader.Value()).To(Equal(segment.SegmentReaderValue{
						SequenceNumber: 0,
						Data:           []byte("foo"),
					}))

					Expect(reader.Next()).To(BeTrue())
					Expect(reader.Err()).To(Succeed())
					Expect(reader.Value()).To(Equal(segment.SegmentReaderValue{
						SequenceNumber: 1,
						Data:           []byte("bar"),
					}))

					Expect(reader.Next()).To(BeTrue())
					Expect(reader.Err()).To(Succeed())
					Expect(reader.Value()).To(Equal(segment.SegmentReaderValue{
						SequenceNumber: 2,
						Data:           []byte("baz"),
					}))

					Expect(reader.Next()).To(BeFalse())
					Expect(reader.Err()).To(MatchError(io.EOF))
				})

				It("should read a pre-allocated segment file", func() {
					writer, err := segment.CreateSegment(dir, 0, segment.CreateSegmentConfig{
						PreAllocationSize:   segment.DefaultPreAllocationSize,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(writer.Close()).To(Succeed())

					reader, err := segment.OpenSegment(dir, 0)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(reader.Close()).To(Succeed())
					}()

					Expect(reader.Next()).To(BeFalse())
					Expect(reader.Err()).ToNot(MatchError(io.EOF))
					Expect(reader.Err()).To(MatchError(segment.ErrEntryNone))
				})
			})
		}
	}

	It("should correctly report sequence numbers", func() {
		var recorder utils.SegmentWriterFileRecorder
		writer, err := segment.NewSegmentWriter(&recorder, segment.NewSegmentWriterConfig{
			Header: encoding.DefaultHeader,
			Offset: encoding.HeaderSize,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := segment.NewSegmentReader(&utils.SegmentReaderFileLoop{
			Data: recorder.Bytes(),
		}, segment.NewSegmentReaderConfig{
			Header:   encoding.DefaultHeader,
			Offset:   encoding.HeaderSize,
			FileSize: math.MaxInt64,
		})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.NextSequenceNumber()).To(Equal(uint64(0)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.NextSequenceNumber()).To(Equal(uint64(1)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.NextSequenceNumber()).To(Equal(uint64(2)))
		Expect(reader.Next()).To(BeTrue())
	})

	It("should correctly report offsets", func() {
		var recorder utils.SegmentWriterFileRecorder
		writer, err := segment.NewSegmentWriter(&recorder, segment.NewSegmentWriterConfig{
			Header: encoding.DefaultHeader,
			Offset: encoding.HeaderSize,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := segment.NewSegmentReader(&utils.SegmentReaderFileLoop{
			Data: recorder.Bytes(),
		}, segment.NewSegmentReaderConfig{
			Header:   encoding.DefaultHeader,
			Offset:   encoding.HeaderSize,
			FileSize: math.MaxInt64,
		})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Offset()).To(Equal(int64(encoding.HeaderSize)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(encoding.HeaderSize + 1*(4+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(encoding.HeaderSize + 2*(4+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(encoding.HeaderSize + 3*(4+3+4))))
	})
})

func BenchmarkSegmentReader_Next(b *testing.B) {
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		for _, entryChecksumType := range encoding.EntryChecksumTypes {
			for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
				data := make([]byte, dataSize*1024)
				var recorder utils.SegmentWriterFileRecorder
				segmentWriter, err := segment.NewSegmentWriter(&recorder, segment.NewSegmentWriterConfig{
					Header: encoding.Header{
						Magic:               encoding.Magic,
						Version:             encoding.HeaderVersion,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					},
				})
				if err != nil {
					b.Fatal(err)
				}

				if _, err := segmentWriter.AppendEntry(data); err != nil {
					b.Fatal(err)
				}

				readerLoop := utils.SegmentReaderFileLoop{
					Data: recorder.Bytes(),
				}
				segmentReader, err := segment.NewSegmentReader(&readerLoop, segment.NewSegmentReaderConfig{
					Header:   segmentWriter.Header(),
					FileSize: math.MaxInt64,
				})
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
