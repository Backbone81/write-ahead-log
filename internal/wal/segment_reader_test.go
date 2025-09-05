package wal_test

import (
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("SegmentReader", func() {
	for _, entryLengthEncoding := range wal.EntryLengthEncodings {
		for _, entryChecksumType := range wal.EntryChecksumTypes {
			var mutex sync.Mutex
			for _, syncPolicy := range []wal.SyncPolicy{
				wal.NewSyncPolicyNone(),
				wal.NewSyncPolicyImmediate(),
				wal.NewSyncPolicyPeriodic(10, time.Millisecond, &mutex),
				wal.NewSyncPolicyGrouped(time.Millisecond, &mutex),
			} {
				Context(fmt.Sprintf("With length encoding %s and entry checksum %s through sync policy %s", entryLengthEncoding, entryChecksumType, syncPolicy), func() {
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
						writer, err := wal.CreateSegment(dir, 0, wal.CreateSegmentConfig{
							PreAllocationSize:   0,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
							SyncPolicy:          syncPolicy,
						})
						Expect(err).ToNot(HaveOccurred())
						Expect(writer.Close()).To(Succeed())

						reader, err := wal.OpenSegment(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						defer func() {
							Expect(reader.Close()).To(Succeed())
						}()

						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).To(MatchError(io.EOF))
					})

					It("should read a full segment file", func() {
						writer, err := wal.CreateSegment(dir, 0, wal.CreateSegmentConfig{
							PreAllocationSize:   0,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
							SyncPolicy:          syncPolicy,
						})
						Expect(err).ToNot(HaveOccurred())
						mutex.Lock()
						Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
						Expect(writer.AppendEntry([]byte("bar"))).Error().ToNot(HaveOccurred())
						Expect(writer.AppendEntry([]byte("baz"))).Error().ToNot(HaveOccurred())
						mutex.Unlock()
						Expect(writer.Close()).To(Succeed())

						reader, err := wal.OpenSegment(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						defer func() {
							Expect(reader.Close()).To(Succeed())
						}()

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Err()).To(Succeed())
						Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
							SequenceNumber: 0,
							Data:           []byte("foo"),
						}))

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Err()).To(Succeed())
						Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
							SequenceNumber: 1,
							Data:           []byte("bar"),
						}))

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Err()).To(Succeed())
						Expect(reader.Value()).To(Equal(wal.SegmentReaderValue{
							SequenceNumber: 2,
							Data:           []byte("baz"),
						}))

						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).To(MatchError(io.EOF))
					})

					It("should read a pre-allocated segment file", func() {
						writer, err := wal.CreateSegment(dir, 0, wal.CreateSegmentConfig{
							PreAllocationSize:   wal.DefaultSegmentSize,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
							SyncPolicy:          syncPolicy,
						})
						Expect(err).ToNot(HaveOccurred())
						Expect(writer.Close()).To(Succeed())

						reader, err := wal.OpenSegment(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						defer func() {
							Expect(reader.Close()).To(Succeed())
						}()

						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).ToNot(MatchError(io.EOF))
						Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))
					})
				})
			}
		}
	}

	It("should correctly report sequence numbers", func() {
		var recorder SegmentWriterFileRecorder
		writer, err := wal.NewSegmentWriter(&recorder, wal.NewSegmentWriterConfig{
			Header:     wal.DefaultHeader,
			Offset:     wal.HeaderSize,
			SyncPolicy: wal.NewSyncPolicyNone(),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.NewSegmentReader(&SegmentReaderFileLoop{
			Data: recorder.Bytes(),
		}, wal.NewSegmentReaderConfig{
			Header:   wal.DefaultHeader,
			Offset:   wal.HeaderSize,
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
		var recorder SegmentWriterFileRecorder
		writer, err := wal.NewSegmentWriter(&recorder, wal.NewSegmentWriterConfig{
			Header:     wal.DefaultHeader,
			Offset:     wal.HeaderSize,
			SyncPolicy: wal.NewSyncPolicyNone(),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.NewSegmentReader(&SegmentReaderFileLoop{
			Data: recorder.Bytes(),
		}, wal.NewSegmentReaderConfig{
			Header:   wal.DefaultHeader,
			Offset:   wal.HeaderSize,
			FileSize: math.MaxInt64,
		})
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
				var recorder SegmentWriterFileRecorder
				segmentWriter, err := wal.NewSegmentWriter(&recorder, wal.NewSegmentWriterConfig{
					Header: wal.Header{
						Magic:               wal.Magic,
						Version:             wal.HeaderVersion,
						EntryLengthEncoding: entryLengthEncoding,
						EntryChecksumType:   entryChecksumType,
					},
					SyncPolicy: wal.NewSyncPolicyNone(),
				})
				if err != nil {
					b.Fatal(err)
				}

				if _, err := segmentWriter.AppendEntry(data); err != nil {
					b.Fatal(err)
				}

				readerLoop := SegmentReaderFileLoop{
					Data: recorder.Bytes(),
				}
				segmentReader, err := wal.NewSegmentReader(&readerLoop, wal.NewSegmentReaderConfig{
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
