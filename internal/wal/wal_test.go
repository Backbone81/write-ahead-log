package wal_test

import (
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/encoding"
	"write-ahead-log/internal/segment"
	"write-ahead-log/internal/wal"
)

var _ = Describe("WAL", func() {
	Context("With default length encoding and default entry checksum through default sync policy", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "test-wal-*")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("should init, write entries and read those entries back again", func() {
			By("initialize WAL")
			Expect(wal.Init(dir)).To(Succeed())

			By("move to end of WAL")
			reader, err := wal.NewReader(dir, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader.Header().FirstSequenceNumber).To(Equal(uint64(0)))
			Expect(reader.Header().EntryLengthEncoding).To(Equal(encoding.DefaultEntryLengthEncoding))
			Expect(reader.Header().EntryChecksumType).To(Equal(encoding.DefaultEntryChecksumType))
			Expect(reader.Next()).To(BeFalse())
			Expect(reader.Err()).To(MatchError(segment.ErrEntryNone))

			By("write to WAL")
			writer, err := reader.ToWriter()
			Expect(err).ToNot(HaveOccurred())
			entries := [][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
			}
			Expect(writer.Header().FirstSequenceNumber).To(Equal(uint64(0)))
			Expect(writer.Header().EntryLengthEncoding).To(Equal(encoding.DefaultEntryLengthEncoding))
			Expect(writer.Header().EntryChecksumType).To(Equal(encoding.DefaultEntryChecksumType))
			for _, entry := range entries {
				Expect(writer.AppendEntry(entry)).Error().ToNot(HaveOccurred())
			}
			Expect(writer.Close()).To(Succeed())

			By("re-open WAL and read the written entries")
			reader, err = wal.NewReader(dir, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader.Header().FirstSequenceNumber).To(Equal(uint64(0)))
			Expect(reader.Header().EntryLengthEncoding).To(Equal(encoding.DefaultEntryLengthEncoding))
			Expect(reader.Header().EntryChecksumType).To(Equal(encoding.DefaultEntryChecksumType))
			for i, entry := range entries {
				Expect(reader.Next()).To(BeTrue())
				Expect(reader.Value().Data).To(Equal(entry))
				Expect(reader.Value().SequenceNumber).To(Equal(uint64(i)))
			}
			Expect(reader.Next()).To(BeFalse())
			Expect(reader.Err()).To(MatchError(segment.ErrEntryNone))
		})
	})

	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		for _, entryChecksumType := range encoding.EntryChecksumTypes {
			for syncPolicyName, syncPolicy := range map[string]wal.WriterOption{
				"none":      wal.WithSyncPolicyNone(),
				"immediate": wal.WithSyncPolicyImmediate(),
				"periodic":  wal.WithSyncPolicyPeriodic(10, time.Millisecond),
				"grouped":   wal.WithSyncPolicyGrouped(time.Millisecond),
			} {
				Context(fmt.Sprintf("With length encoding %s and entry checksum %s through sync policy %s", entryLengthEncoding, entryChecksumType, syncPolicyName), func() {
					var dir string

					BeforeEach(func() {
						var err error
						dir, err = os.MkdirTemp("", "test-wal-*")
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(dir)).To(Succeed())
					})

					It("should init, write entries and read those entries back again", func() {
						By("initialize WAL")
						Expect(wal.Init(dir, wal.WithEntryLengthEncoding(entryLengthEncoding), wal.WithEntryChecksumType(entryChecksumType))).To(Succeed())

						By("move to end of WAL")
						reader, err := wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Header().FirstSequenceNumber).To(Equal(uint64(0)))
						Expect(reader.Header().EntryLengthEncoding).To(Equal(entryLengthEncoding))
						Expect(reader.Header().EntryChecksumType).To(Equal(entryChecksumType))
						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).To(MatchError(segment.ErrEntryNone))

						By("write to WAL")
						writer, err := reader.ToWriter(syncPolicy)
						Expect(err).ToNot(HaveOccurred())
						Expect(writer.Header().FirstSequenceNumber).To(Equal(uint64(0)))
						Expect(writer.Header().EntryLengthEncoding).To(Equal(entryLengthEncoding))
						Expect(writer.Header().EntryChecksumType).To(Equal(entryChecksumType))
						entries := [][]byte{
							[]byte("foo"),
							[]byte("bar"),
							[]byte("baz"),
						}
						for _, entry := range entries {
							Expect(writer.AppendEntry(entry)).Error().ToNot(HaveOccurred())
						}
						Expect(writer.Close()).To(Succeed())

						By("re-open WAL and read the written entries")
						reader, err = wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Header().FirstSequenceNumber).To(Equal(uint64(0)))
						Expect(reader.Header().EntryLengthEncoding).To(Equal(entryLengthEncoding))
						Expect(reader.Header().EntryChecksumType).To(Equal(entryChecksumType))
						for i, entry := range entries {
							Expect(reader.Next()).To(BeTrue())
							Expect(reader.Value().Data).To(Equal(entry))
							Expect(reader.Value().SequenceNumber).To(Equal(uint64(i)))
						}
						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).To(MatchError(segment.ErrEntryNone))
					})

					It("should panic to close the reader when the writer was already created", func() {
						By("initialize WAL")
						Expect(wal.Init(dir, wal.WithEntryLengthEncoding(entryLengthEncoding), wal.WithEntryChecksumType(entryChecksumType))).To(Succeed())

						By("move to end of WAL")
						reader, err := wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Next()).To(BeFalse())

						By("create writer")
						writer, err := reader.ToWriter(syncPolicy)
						Expect(err).ToNot(HaveOccurred())

						Expect(func() {
							_ = reader.Close()
						}).To(Panic())

						Expect(writer.Close()).To(Succeed())
					})

					It("should roll over the segment", func() {
						By("initialize WAL")
						Expect(wal.Init(
							dir,
							wal.WithEntryLengthEncoding(entryLengthEncoding),
							wal.WithEntryChecksumType(entryChecksumType),
							wal.WithPreAllocationSize(512),
						)).To(Succeed())

						By("move to end of WAL")
						reader, err := wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Next()).To(BeFalse())

						By("create writer")
						var rolloverCount int
						writer, err := reader.ToWriter(
							syncPolicy,
							wal.WithMaxSegmentSize(512),
							wal.WithPreAllocationSize(512),
							wal.WithRolloverCallback(func(previousSegment uint64, nextSegment uint64) {
								rolloverCount++
							}),
						)
						Expect(err).ToNot(HaveOccurred())

						initialSegment := writer.FilePath()
						Expect(writer.AppendEntry(make([]byte, 1024))).Error().ToNot(HaveOccurred())
						Expect(writer.FilePath()).To(Equal(initialSegment))

						// The rollover happens on the first write after we are over the size
						Expect(writer.AppendEntry([]byte("bar"))).Error().ToNot(HaveOccurred())
						Expect(writer.FilePath()).ToNot(Equal(initialSegment))
						Expect(rolloverCount).To(Equal(1))

						Expect(writer.Close()).To(Succeed())
					})

					It("should correctly deal with too small of a segment size", func() {
						By("initialize WAL")
						Expect(wal.Init(
							dir,
							wal.WithEntryLengthEncoding(entryLengthEncoding),
							wal.WithEntryChecksumType(entryChecksumType),
							wal.WithPreAllocationSize(0),
						)).To(Succeed())

						By("move to end of WAL")
						reader, err := wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Next()).To(BeFalse())

						By("create writer")
						var rolloverCount int
						writer, err := reader.ToWriter(
							syncPolicy,
							wal.WithMaxSegmentSize(0),
							wal.WithPreAllocationSize(0),
							wal.WithRolloverCallback(func(previousSegment uint64, nextSegment uint64) {
								rolloverCount++
							}),
						)
						Expect(err).ToNot(HaveOccurred())

						initialSegment := writer.FilePath()
						Expect(writer.AppendEntry([]byte("bar"))).Error().ToNot(HaveOccurred())
						Expect(writer.FilePath()).To(Equal(initialSegment))
						Expect(rolloverCount).To(Equal(0))

						Expect(writer.Close()).To(Succeed())
					})

					It("should correctly deal with segment sizes below the pre-allocation size", func() {
						By("initialize WAL")
						Expect(wal.Init(
							dir,
							wal.WithEntryLengthEncoding(entryLengthEncoding),
							wal.WithEntryChecksumType(entryChecksumType),
							wal.WithPreAllocationSize(1024*1024),
						)).To(Succeed())

						By("move to end of WAL")
						reader, err := wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Next()).To(BeFalse())

						By("create writer")
						var rolloverCount int
						writer, err := reader.ToWriter(
							syncPolicy,
							wal.WithMaxSegmentSize(0),
							wal.WithPreAllocationSize(1024*1024),
							wal.WithRolloverCallback(func(previousSegment uint64, nextSegment uint64) {
								rolloverCount++
							}),
						)
						Expect(err).ToNot(HaveOccurred())

						Expect(writer.AppendEntry([]byte("foo"))).Error().ToNot(HaveOccurred())
						Expect(writer.AppendEntry([]byte("bar"))).Error().ToNot(HaveOccurred())
						Expect(writer.AppendEntry([]byte("baz"))).Error().ToNot(HaveOccurred())
						Expect(rolloverCount).To(Equal(2))

						Expect(writer.Close()).To(Succeed())

						By("read back")
						reader, err = wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Value().Data).To(Equal([]byte("foo")))

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Value().Data).To(Equal([]byte("bar")))

						Expect(reader.Next()).To(BeTrue())
						Expect(reader.Value().Data).To(Equal([]byte("baz")))

						Expect(reader.Next()).To(BeFalse())
					})
				})
			}
		}
	}
})

//nolint:gocognit,cyclop
func BenchmarkWriter_AppendEntry_Serial(b *testing.B) {
	for _, entryLengthEncoding := range []encoding.EntryLengthEncoding{encoding.DefaultEntryLengthEncoding} {
		for _, entryChecksumType := range []encoding.EntryChecksumType{encoding.DefaultEntryChecksumType} {
			for syncPolicyName, syncPolicy := range map[string]wal.WriterOption{
				"none":      wal.WithSyncPolicyNone(),
				"immediate": wal.WithSyncPolicyImmediate(),
				"periodic":  wal.WithSyncPolicyPeriodic(100, 10*time.Millisecond),
				"grouped":   wal.WithSyncPolicyGrouped(10 * time.Millisecond),
			} {
				for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
					dir := b.TempDir()
					data := make([]byte, dataSize*1024)
					if err := wal.Init(
						dir,
						wal.WithEntryLengthEncoding(entryLengthEncoding),
						wal.WithEntryChecksumType(entryChecksumType),
					); err != nil {
						b.Fatal(err)
					}
					reader, err := wal.NewReader(dir, 0)
					if err != nil {
						b.Fatal(err)
					}
					reader.Next()
					writer, err := reader.ToWriter(syncPolicy, wal.WithRolloverCallback(func(previousSegment uint64, nextSegment uint64) {
						if err := os.Remove(path.Join(dir, segment.SegmentFileName(previousSegment))); err != nil {
							b.Fatal(err)
						}
					}))
					if err != nil {
						b.Fatal(err)
					}
					b.Run(fmt.Sprintf("%s %s %s %d KB", entryLengthEncoding, entryChecksumType, syncPolicyName, dataSize), func(b *testing.B) {
						for range b.N {
							_, err = writer.AppendEntry(data)
							if err != nil {
								panic(err)
							}
						}
						timeNeeded := b.Elapsed().Seconds()
						dataAppended := b.N * dataSize * 1024
						b.ReportMetric(float64(dataAppended/1024/1024)/timeNeeded, "MB/s")
					})
					if err := writer.Close(); err != nil {
						b.Fatal(err)
					}
				}
			}
		}
	}
}

//nolint:gocognit,cyclop
func BenchmarkWriter_AppendEntry_Concurrently(b *testing.B) {
	for _, entryLengthEncoding := range []encoding.EntryLengthEncoding{encoding.DefaultEntryLengthEncoding} {
		for _, entryChecksumType := range []encoding.EntryChecksumType{encoding.DefaultEntryChecksumType} {
			for syncPolicyName, syncPolicy := range map[string]wal.WriterOption{
				"none":      wal.WithSyncPolicyNone(),
				"immediate": wal.WithSyncPolicyImmediate(),
				"periodic":  wal.WithSyncPolicyPeriodic(100, 10*time.Millisecond),
				"grouped":   wal.WithSyncPolicyGrouped(10 * time.Millisecond),
			} {
				for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
					dir := b.TempDir()
					data := make([]byte, dataSize*1024)
					if err := wal.Init(
						dir,
						wal.WithEntryLengthEncoding(entryLengthEncoding),
						wal.WithEntryChecksumType(entryChecksumType),
					); err != nil {
						b.Fatal(err)
					}
					reader, err := wal.NewReader(dir, 0)
					if err != nil {
						b.Fatal(err)
					}
					reader.Next()
					writer, err := reader.ToWriter(syncPolicy, wal.WithRolloverCallback(func(previousSegment uint64, nextSegment uint64) {
						if err := os.Remove(path.Join(dir, segment.SegmentFileName(previousSegment))); err != nil {
							b.Fatal(err)
						}
					}))
					if err != nil {
						b.Fatal(err)
					}
					b.Run(fmt.Sprintf("%s %s %s %d KB", entryLengthEncoding, entryChecksumType, syncPolicyName, dataSize), func(b *testing.B) {
						var wg sync.WaitGroup
						wg.Add(b.N)
						for range b.N {
							go func() {
								defer wg.Done()
								_, err = writer.AppendEntry(data)
								if err != nil {
									panic(err)
								}
							}()
						}
						wg.Wait()
						timeNeeded := b.Elapsed().Seconds()
						dataAppended := b.N * dataSize * 1024
						b.ReportMetric(float64(dataAppended/1024/1024)/timeNeeded, "MB/s")
					})
					if err := writer.Close(); err != nil {
						b.Fatal(err)
					}
				}
			}
		}
	}
}
