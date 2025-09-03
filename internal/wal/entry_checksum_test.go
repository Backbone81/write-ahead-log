package wal_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("EntryChecksum", func() {
	DescribeTable("Writing entry checksums",
		func(entryChecksumType wal.EntryChecksumType, wantBytes int) {
			writer, err := wal.GetEntryChecksumWriter(entryChecksumType)
			Expect(err).ToNot(HaveOccurred())

			var output bytes.Buffer
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, 1024)
			Expect(writer(&output, buffer[:], data)).To(Succeed())
			Expect(output.Len()).To(Equal(wantBytes))
		},
		Entry("When using CRC32", wal.EntryChecksumTypeCrc32, 4),
		Entry("When using CRC64", wal.EntryChecksumTypeCrc64, 8),
	)

	DescribeTable("Reading entry checksums",
		func(entryChecksumType wal.EntryChecksumType) {
			writer, err := wal.GetEntryChecksumWriter(entryChecksumType)
			Expect(err).ToNot(HaveOccurred())

			reader, err := wal.GetEntryChecksumReader(entryChecksumType)
			Expect(err).ToNot(HaveOccurred())

			var output bytes.Buffer
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, 1024)
			Expect(writer(&output, buffer[:], data)).To(Succeed())
			Expect(reader(&output, buffer[:], data)).Error().ToNot(HaveOccurred())
		},
		Entry("When using CRC32", wal.EntryChecksumTypeCrc32),
		Entry("When using CRC64", wal.EntryChecksumTypeCrc64),
	)
})

func BenchmarkEntryChecksumWriter(b *testing.B) {
	var buffer [wal.MaxChecksumBufferLen]byte
	for _, entryChecksumType := range wal.EntryChecksumTypes {
		writer, err := wal.GetEntryChecksumWriter(entryChecksumType)
		if err != nil {
			b.Fatal(err)
		}
		for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
			data := make([]byte, dataSize*1024)
			b.Run(fmt.Sprintf("%s on %d KB", entryChecksumType, dataSize), func(b *testing.B) {
				for b.Loop() {
					if err := writer(io.Discard, buffer[:], data); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkEntryChecksumReader(b *testing.B) {
	var buffer [wal.MaxChecksumBufferLen]byte
	var checksum [wal.MaxChecksumBufferLen]byte
	for _, entryChecksumType := range wal.EntryChecksumTypes {
		writer, err := wal.GetEntryChecksumWriter(entryChecksumType)
		if err != nil {
			b.Fatal(err)
		}
		reader, err := wal.GetEntryChecksumReader(entryChecksumType)
		if err != nil {
			b.Fatal(err)
		}
		for _, dataSize := range []int{0, 1, 2, 4, 8, 16} {
			data := make([]byte, dataSize*1024)
			if err := writer(bytes.NewBuffer(checksum[:0]), buffer[:], data); err != nil {
				b.Fatal(err)
			}
			checksumReader := bytes.NewReader(checksum[:])
			checksumReaderBackup := *checksumReader
			// We do the interface conversion outside the loop to avoid the memory allocation of the interface
			// conversion to skew the measurements.
			var checksumIoReader io.Reader = checksumReader
			b.Run(fmt.Sprintf("%s on %d KB", entryChecksumType, dataSize), func(b *testing.B) {
				for b.Loop() {
					*checksumReader = checksumReaderBackup
					if _, err := reader(checksumIoReader, buffer[:], data); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
