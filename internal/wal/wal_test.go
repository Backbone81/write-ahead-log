package wal_test

import (
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("WAL", func() {
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
						for _, entry := range entries {
							Expect(writer.AppendEntry(entry)).To(Succeed())
						}
						Expect(writer.Close()).To(Succeed())
						writer.Unlock()

						By("re-open WAL and read the written entries")
						reader, err = wal.NewReader(dir, 0)
						Expect(err).ToNot(HaveOccurred())
						for i, entry := range entries {
							Expect(reader.Next()).To(BeTrue())
							Expect(reader.Value().Data).To(Equal(entry))
							Expect(reader.Value().SequenceNumber).To(Equal(uint64(i)))
						}
						Expect(reader.Next()).To(BeFalse())
						Expect(reader.Err()).To(MatchError(wal.ErrEntryNone))
					})
				})
			}
		}
	}
})
