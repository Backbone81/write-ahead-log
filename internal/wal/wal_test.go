package wal_test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
			Expect(reader.Header().EntryLengthEncoding).To(Equal(wal.DefaultEntryLengthEncoding))
			Expect(reader.Header().EntryChecksumType).To(Equal(wal.DefaultEntryChecksumType))
			Expect(reader.Next()).To(BeFalse())
			Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))

			By("write to WAL")
			writer, err := reader.ToWriter()
			Expect(err).ToNot(HaveOccurred())
			entries := [][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
			}
			writer.Lock()
			Expect(writer.Header().FirstSequenceNumber).To(Equal(uint64(0)))
			Expect(writer.Header().EntryLengthEncoding).To(Equal(wal.DefaultEntryLengthEncoding))
			Expect(writer.Header().EntryChecksumType).To(Equal(wal.DefaultEntryChecksumType))
			for _, entry := range entries {
				Expect(writer.AppendEntry(entry)).To(Succeed())
			}
			Expect(writer.Close()).To(Succeed())
			writer.Unlock()

			By("re-open WAL and read the written entries")
			reader, err = wal.NewReader(dir, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader.Header().FirstSequenceNumber).To(Equal(uint64(0)))
			Expect(reader.Header().EntryLengthEncoding).To(Equal(wal.DefaultEntryLengthEncoding))
			Expect(reader.Header().EntryChecksumType).To(Equal(wal.DefaultEntryChecksumType))
			for i, entry := range entries {
				Expect(reader.Next()).To(BeTrue())
				Expect(reader.Value().Data).To(Equal(entry))
				Expect(reader.Value().SequenceNumber).To(Equal(uint64(i)))
			}
			Expect(reader.Next()).To(BeFalse())
			Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))
		})
	})

	for _, entryLengthEncoding := range wal.EntryLengthEncodings {
		for _, entryChecksumType := range wal.EntryChecksumTypes {
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
						Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))

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
						writer.Lock()
						for _, entry := range entries {
							Expect(writer.AppendEntry(entry)).To(Succeed())
						}
						Expect(writer.Close()).To(Succeed())
						writer.Unlock()

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
						Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))
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

						writer.Lock()
						Expect(writer.Close()).To(Succeed())
						writer.Unlock()
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

						writer.Lock()
						initialSegment := writer.FilePath()
						Expect(writer.AppendEntry(make([]byte, 1024))).To(Succeed())
						Expect(writer.FilePath()).To(Equal(initialSegment))

						// The rollover happens on the first write after we are over the size
						Expect(writer.AppendEntry([]byte("bar"))).To(Succeed())
						Expect(writer.FilePath()).ToNot(Equal(initialSegment))
						Expect(rolloverCount).To(Equal(1))

						Expect(writer.Close()).To(Succeed())
						writer.Unlock()
					})
				})
			}
		}
	}
})

func BenchmarkWriter_AppendEntry(b *testing.B) {
	for _, entryLengthEncoding := range []wal.EntryLengthEncoding{wal.DefaultEntryLengthEncoding} {
		for _, entryChecksumType := range []wal.EntryChecksumType{wal.DefaultEntryChecksumType} {
			for syncPolicyName, syncPolicy := range map[string]wal.WriterOption{
				//"none":      wal.WithSyncPolicyNone(),
				//"immediate": wal.WithSyncPolicyImmediate(),
				//"periodic":  wal.WithSyncPolicyPeriodic(10, time.Millisecond),
				"grouped": wal.WithSyncPolicyGrouped(time.Millisecond),
			} {
				for _, dataSize := range []int{4} {
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
						//if err := os.Remove(path.Join(dir, wal.SegmentFileName(previousSegment))); err != nil {
						//	b.Fatal(err)
						//}
					}))
					if err != nil {
						b.Fatal(err)
					}
					b.Run(fmt.Sprintf("%s %s %s %d KB", entryLengthEncoding, entryChecksumType, syncPolicyName, dataSize), func(b *testing.B) {
						var wg sync.WaitGroup
						count := 5000
						wg.Add(count)
						for range count {
							go func() {
								writer.Lock()
								_ = writer.AppendEntry(data)
								writer.Unlock()
								wg.Done()
							}()
						}
						wg.Wait()
					})
					writer.Lock()
					if err := writer.Close(); err != nil {
						b.Fatal(err)
					}
					writer.Unlock()
				}
			}
		}
	}
}
