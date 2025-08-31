package wal_test

import (
	"fmt"
	"io"
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

		reader, err := wal.NewSegmentReader(dir, 7)
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

		reader, err := wal.NewSegmentReader(dir, 7)
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

	PIt("should read a pre-allocated segment file", func() {
		writer, err := wal.CreateSegment(dir, 7, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		Expect(err).ToNot(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		reader, err := wal.NewSegmentReader(dir, 7)
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

		reader, err := wal.NewSegmentReader(dir, 7)
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

		reader, err := wal.NewSegmentReader(dir, 7)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(reader.Close()).To(Succeed())
		}()

		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize)))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 1*(8+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 2*(8+3+4))))
		Expect(reader.Next()).To(BeTrue())
		Expect(reader.Offset()).To(Equal(int64(wal.HeaderSize + 3*(8+3+4))))
	})
})

func BenchmarkSegmentReader_Next_1000x(b *testing.B) {
	for _, i := range []int{0, 1, 2, 4, 8} {
		dir := b.TempDir()
		writer, err := wal.CreateSegment(dir, 0, wal.DefaultSegmentSize, &wal.SyncPolicyNone{})
		if err != nil {
			b.Fatal(err)
		}

		data := make([]byte, i*1024)

		for range 1000 {
			if err := writer.AppendEntry(data); err != nil {
				b.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("%d KB data", i), func(b *testing.B) {
			for range b.N {
				readSegmentWith1000Entries(b, dir)
			}
		})
	}
}

func readSegmentWith1000Entries(b *testing.B, dir string) {
	b.Helper()
	b.StopTimer()

	reader, err := wal.NewSegmentReader(dir, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			b.Fatal(err)
		}
	}()

	b.StartTimer()
	for range 1000 {
		if !reader.Next() {
			b.Fatal("could not read 1000 entries from segment")
		}
	}
	b.StopTimer()
}
