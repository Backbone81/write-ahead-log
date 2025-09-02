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
		func(writer wal.EntryChecksumWriter, wantBytes int) {
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, 1024)
			var output bytes.Buffer
			Expect(writer(&output, buffer[:], data)).To(Succeed())
			Expect(output.Len()).To(Equal(wantBytes))
		},
		Entry("When using CRC32", wal.WriteEntryChecksumCrc32, 4),
		Entry("When using CRC64", wal.WriteEntryChecksumCrc64, 8),
	)

	DescribeTable("Reading entry checksums",
		func(writer wal.EntryChecksumWriter, reader wal.EntryChecksumReader) {
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, 1024)
			var output bytes.Buffer
			Expect(writer(&output, buffer[:], data)).To(Succeed())
			Expect(reader(&output, buffer[:], data)).Error().ToNot(HaveOccurred())
		},
		Entry("When using CRC32", wal.WriteEntryChecksumCrc32, wal.ReadEntryChecksumCrc32),
		Entry("When using CRC64", wal.WriteEntryChecksumCrc64, wal.ReadEntryChecksumCrc64),
	)
})

func BenchmarkEntryChecksumWriter(b *testing.B) {
	entryChecksumWriters := []struct {
		name   string
		writer wal.EntryChecksumWriter
	}{
		{
			name:   "CRC32",
			writer: wal.WriteEntryChecksumCrc32,
		},
		{
			name:   "CRC64",
			writer: wal.WriteEntryChecksumCrc64,
		},
	}
	for _, entryChecksumWriter := range entryChecksumWriters {
		for _, i := range []int{0, 1, 2, 4, 8, 16} {
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, i*1024)
			b.Run(fmt.Sprintf("%s on %d KB data", entryChecksumWriter.name, i), func(b *testing.B) {
				for b.Loop() {
					if err := entryChecksumWriter.writer(io.Discard, buffer[:], data); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkEntryChecksumReader(b *testing.B) {
	entryChecksumReaders := []struct {
		name   string
		writer wal.EntryChecksumWriter
		reader wal.EntryChecksumReader
	}{
		{
			name:   "CRC32",
			writer: wal.WriteEntryChecksumCrc32,
			reader: wal.ReadEntryChecksumCrc32,
		},
		{
			name:   "CRC64",
			writer: wal.WriteEntryChecksumCrc64,
			reader: wal.ReadEntryChecksumCrc64,
		},
	}
	for _, entryChecksumReader := range entryChecksumReaders {
		for _, i := range []int{0, 1, 2, 4, 8, 16} {
			var buffer [wal.MaxChecksumBufferLen]byte
			data := make([]byte, i*1024)
			var checksum [wal.MaxChecksumBufferLen]byte
			if err := entryChecksumReader.writer(bytes.NewBuffer(checksum[:0]), buffer[:], data); err != nil {
				b.Fatal(err)
			}
			checksumReader := bytes.NewReader(checksum[:])
			checksumReaderBackup := *checksumReader
			// We do the interface conversion outside the loop to avoid the memory allocation of the interface
			// conversion to skew the measurements.
			var checksumIoReader io.Reader = checksumReader
			b.Run(fmt.Sprintf("%s on %d KB data", entryChecksumReader.name, i), func(b *testing.B) {
				for b.Loop() {
					*checksumReader = checksumReaderBackup
					if _, err := entryChecksumReader.reader(checksumIoReader, buffer[:], data); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
