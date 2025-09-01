package wal_test

import (
	"bytes"
	"io"
	"math"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("EntryLength", func() {
	DescribeTable("Writing entry lengths",
		func(writer wal.EntryLengthWriter, value uint64, wantBytes int) {
			var buffer [wal.MaxLengthBufferLen]byte
			var output bytes.Buffer
			Expect(writer(&output, buffer[:], value)).To(Succeed())
			Expect(output.Len()).To(Equal(wantBytes))
		},
		Entry("When using uint16", wal.WriteEntryLengthUint16, uint64(1), 2),
		Entry("When using uint32", wal.WriteEntryLengthUint32, uint64(1), 4),
		Entry("When using uint64", wal.WriteEntryLengthUint64, uint64(1), 8),
		Entry("When using uvarint low", wal.WriteEntryLengthUvarint, uint64(1), 1),
		Entry("When using uvarint MaxUint8", wal.WriteEntryLengthUvarint, uint64(math.MaxUint8), 2),
		Entry("When using uvarint MaxUint16", wal.WriteEntryLengthUvarint, uint64(math.MaxUint16), 3),
		Entry("When using uvarint MaxUint32", wal.WriteEntryLengthUvarint, uint64(math.MaxUint32), 5),
		Entry("When using uvarint MaxUint64", wal.WriteEntryLengthUvarint, uint64(math.MaxUint64), 10),
	)

	DescribeTable("Reading entry lengths",
		func(writer wal.EntryLengthWriter, reader wal.EntryLengthReader, value uint64) {
			var buffer [wal.MaxLengthBufferLen]byte
			var output bytes.Buffer
			Expect(writer(&output, buffer[:], value)).To(Succeed())
			readValue, err := reader(&output, buffer[:])
			Expect(err).ToNot(HaveOccurred())
			Expect(readValue).To(Equal(value))
		},
		Entry("When using uint16", wal.WriteEntryLengthUint16, wal.ReadEntryLengthUint16, uint64(1)),
		Entry("When using uint32", wal.WriteEntryLengthUint32, wal.ReadEntryLengthUint32, uint64(1)),
		Entry("When using uint64", wal.WriteEntryLengthUint64, wal.ReadEntryLengthUint64, uint64(1)),
		Entry("When using uvarint low", wal.WriteEntryLengthUvarint, wal.ReadEntryLengthUvarint, uint64(1)),
		Entry("When using uvarint MaxUint8", wal.WriteEntryLengthUvarint, wal.ReadEntryLengthUvarint, uint64(math.MaxUint8)),
		Entry("When using uvarint MaxUint16", wal.WriteEntryLengthUvarint, wal.ReadEntryLengthUvarint, uint64(math.MaxUint16)),
		Entry("When using uvarint MaxUint32", wal.WriteEntryLengthUvarint, wal.ReadEntryLengthUvarint, uint64(math.MaxUint32)),
		Entry("When using uvarint MaxUint64", wal.WriteEntryLengthUvarint, wal.ReadEntryLengthUvarint, uint64(math.MaxUint64)),
	)
})

func BenchmarkEntryLengthWriter(b *testing.B) {
	entryLengthWriters := []struct {
		name   string
		writer wal.EntryLengthWriter
	}{
		{
			name:   "uint16",
			writer: wal.WriteEntryLengthUint16,
		},
		{
			name:   "uint32",
			writer: wal.WriteEntryLengthUint32,
		},
		{
			name:   "uint64",
			writer: wal.WriteEntryLengthUint64,
		},
		{
			name:   "uvarint",
			writer: wal.WriteEntryLengthUvarint,
		},
	}
	for _, entryLengthWriter := range entryLengthWriters {
		var buffer [wal.MaxLengthBufferLen]byte
		b.Run(entryLengthWriter.name, func(b *testing.B) {
			for b.Loop() {
				if err := entryLengthWriter.writer(io.Discard, buffer[:], 1024); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEntryLengthReader(b *testing.B) {
	entryLengthReaders := []struct {
		name   string
		writer wal.EntryLengthWriter
		reader wal.EntryLengthReader
	}{
		{
			name:   "uint16",
			writer: wal.WriteEntryLengthUint16,
			reader: wal.ReadEntryLengthUint16,
		},
		{
			name:   "uint32",
			writer: wal.WriteEntryLengthUint32,
			reader: wal.ReadEntryLengthUint32,
		},
		{
			name:   "uint64",
			writer: wal.WriteEntryLengthUint64,
			reader: wal.ReadEntryLengthUint64,
		},
		{
			name:   "uvarint",
			writer: wal.WriteEntryLengthUvarint,
			reader: wal.ReadEntryLengthUvarint,
		},
	}
	for _, entryLengthReader := range entryLengthReaders {
		var buffer [wal.MaxLengthBufferLen]byte
		var output bytes.Buffer
		if err := entryLengthReader.writer(&output, buffer[:], 1024); err != nil {
			b.Fatal(err)
		}
		outputBackup := output
		// We do the interface conversion outside the loop to not have the allocation of the interface conversion
		// skew the measurements.
		var outputAsIoReader io.Reader = &output
		b.Run(entryLengthReader.name, func(b *testing.B) {
			for b.Loop() {
				output = outputBackup
				if _, err := entryLengthReader.reader(outputAsIoReader, buffer[:]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
