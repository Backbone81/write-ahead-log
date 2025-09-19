package segment_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/write-ahead-log/internal/encoding"
	"github.com/backbone81/write-ahead-log/internal/segment"
	"github.com/backbone81/write-ahead-log/internal/utils"
)

var _ = Describe("SegmentWriter", func() {
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		for _, entryChecksumType := range encoding.EntryChecksumTypes {
			Context(fmt.Sprintf("With length encoding %s and entry checksum %s", entryLengthEncoding, entryChecksumType), func() {
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

					writer, err := segment.CreateSegment(dir, 0, segment.CreateSegmentConfig{
						PreAllocationSize:   segment.DefaultPreAllocationSize,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					})
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(writer.Close()).To(Succeed())
					}()

					entriesAfter, err := os.ReadDir(dir)
					Expect(err).ToNot(HaveOccurred())
					Expect(entriesAfter).To(HaveLen(len(entriesBefore) + 1))
				})

				It("should write to the segment file", func() {
					writer, err := segment.CreateSegment(dir, 0, segment.CreateSegmentConfig{
						PreAllocationSize:   segment.DefaultPreAllocationSize,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					})
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(writer.Close()).To(Succeed())
					}()

					for range 10 {
						var data [1024]byte
						Expect(rand.Read(data[:])).Error().ToNot(HaveOccurred())
						Expect(writer.AppendEntry(data[:])).Error().ToNot(HaveOccurred())
					}
				})
			})
		}
	}

	It("should correctly report sequence numbers", func() {
		writer, err := segment.NewSegmentWriter(&utils.SegmentWriterFileDiscard{}, segment.NewSegmentWriterConfig{
			Header: encoding.DefaultHeader,
			Offset: encoding.HeaderSize,
		})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.NextSequenceNumber()).To(Equal(uint64(0)))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(1)))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(2)))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.NextSequenceNumber()).To(Equal(uint64(3)))
	})

	It("should correctly report offsets", func() {
		writer, err := segment.NewSegmentWriter(&utils.SegmentWriterFileDiscard{}, segment.NewSegmentWriterConfig{
			Header: encoding.DefaultHeader,
			Offset: encoding.HeaderSize,
		})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.Offset()).To(Equal(int64(encoding.HeaderSize)))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(encoding.HeaderSize + 1*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(encoding.HeaderSize + 2*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(encoding.HeaderSize + 3*(4+3+4))))
	})
})

func BenchmarkSegmentWriter_AppendEntry(b *testing.B) {
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		for _, entryChecksumType := range encoding.EntryChecksumTypes {
			for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
				data := make([]byte, dataSize*1024)
				segmentWriter, err := segment.NewSegmentWriter(&utils.SegmentWriterFileDiscard{}, segment.NewSegmentWriterConfig{
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
				b.Run(fmt.Sprintf("%s %s %d KB", entryLengthEncoding, entryChecksumType, dataSize), func(b *testing.B) {
					for b.Loop() {
						if _, err := segmentWriter.AppendEntry(data); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
