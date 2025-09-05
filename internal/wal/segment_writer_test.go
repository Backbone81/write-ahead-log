package wal_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("SegmentWriter", func() {
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
						dir, err = os.MkdirTemp("", "test-segment-writer-*")
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(dir)).To(Succeed())
					})

					It("should create a new segment file", func() {
						entriesBefore, err := os.ReadDir(dir)
						Expect(err).ToNot(HaveOccurred())

						writer, err := wal.CreateSegment(dir, 0, wal.CreateSegmentConfig{
							PreAllocationSize:   wal.DefaultSegmentSize,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
							SyncPolicy:          syncPolicy,
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
						writer, err := wal.CreateSegment(dir, 0, wal.CreateSegmentConfig{
							PreAllocationSize:   wal.DefaultSegmentSize,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
							SyncPolicy:          syncPolicy,
						})
						Expect(err).ToNot(HaveOccurred())
						defer func() {
							Expect(writer.Close()).To(Succeed())
						}()

						for range 10 {
							var data [1024]byte
							Expect(rand.Read(data[:])).Error().ToNot(HaveOccurred())
							mutex.Lock()
							Expect(writer.AppendEntry(data[:])).Error().ToNot(HaveOccurred())
							mutex.Unlock()
						}
					})
				})
			}
		}
	}

	It("should correctly report sequence numbers", func() {
		writer, err := wal.NewSegmentWriter(&SegmentWriterFileDiscard{}, wal.NewSegmentWriterConfig{
			Header:     wal.DefaultHeader,
			Offset:     wal.HeaderSize,
			SyncPolicy: wal.NewSyncPolicyNone(),
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
		writer, err := wal.NewSegmentWriter(&SegmentWriterFileDiscard{}, wal.NewSegmentWriterConfig{
			Header:     wal.DefaultHeader,
			Offset:     wal.HeaderSize,
			SyncPolicy: wal.NewSyncPolicyNone(),
		})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(writer.Close()).To(Succeed())
		}()

		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize)))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 1*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 2*(4+3+4))))
		Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
		Expect(writer.Offset()).To(Equal(int64(wal.HeaderSize + 3*(4+3+4))))
	})
})

func BenchmarkSegmentWriter_AppendEntry(b *testing.B) {
	for _, entryLengthEncoding := range wal.EntryLengthEncodings {
		for _, entryChecksumType := range wal.EntryChecksumTypes {
			var mutex sync.Mutex
			for _, syncPolicy := range []wal.SyncPolicy{
				wal.NewSyncPolicyNone(),
				wal.NewSyncPolicyImmediate(),
				wal.NewSyncPolicyPeriodic(10, time.Millisecond, &mutex),
				wal.NewSyncPolicyGrouped(time.Millisecond, &mutex),
			} {
				for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
					data := make([]byte, dataSize*1024)
					segmentWriter, err := wal.NewSegmentWriter(&SegmentWriterFileDiscard{}, wal.NewSegmentWriterConfig{
						Header: wal.Header{
							Magic:               wal.Magic,
							Version:             wal.HeaderVersion,
							EntryLengthEncoding: entryLengthEncoding,
							EntryChecksumType:   entryChecksumType,
						},
						SyncPolicy: syncPolicy,
					})
					if err != nil {
						b.Fatal(err)
					}
					b.Run(fmt.Sprintf("%s %s %s %d KB", entryLengthEncoding, entryChecksumType, syncPolicy, dataSize), func(b *testing.B) {
						mutex.Lock()
						defer mutex.Unlock()
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
}
