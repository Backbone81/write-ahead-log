package encoding_test

import (
	"bytes"
	"io"
	"math"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/encoding"
)

var _ = Describe("EntryLength", func() {
	DescribeTable("Writing entry lengths",
		func(entryLengthEncoding encoding.EntryLengthEncoding, value uint64, wantBytes int) {
			writer, err := encoding.GetEntryLengthWriter(entryLengthEncoding)
			Expect(err).ToNot(HaveOccurred())

			var output bytes.Buffer
			var buffer [encoding.MaxLengthBufferLen]byte
			Expect(writer(&output, buffer[:], value)).To(Succeed())
			Expect(output.Len()).To(Equal(wantBytes))
		},
		Entry("When using uint16", encoding.EntryLengthEncodingUint16, uint64(1), 2),
		Entry("When using uint32", encoding.EntryLengthEncodingUint32, uint64(1), 4),
		Entry("When using uint64", encoding.EntryLengthEncodingUint64, uint64(1), 8),
		Entry("When using uvarint low", encoding.EntryLengthEncodingUvarint, uint64(1), 1),
		Entry("When using uvarint MaxUint8", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint8), 2),
		Entry("When using uvarint MaxUint16", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint16), 3),
		Entry("When using uvarint MaxUint32", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint32), 5),
		Entry("When using uvarint MaxUint64", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint64), 10),
	)

	DescribeTable("Reading entry lengths",
		func(entryLengthEncoding encoding.EntryLengthEncoding, value uint64) {
			writer, err := encoding.GetEntryLengthWriter(entryLengthEncoding)
			Expect(err).ToNot(HaveOccurred())

			reader, err := encoding.GetEntryLengthReader(entryLengthEncoding)
			Expect(err).ToNot(HaveOccurred())

			var output bytes.Buffer
			var buffer [encoding.MaxLengthBufferLen]byte
			Expect(writer(&output, buffer[:], value)).To(Succeed())
			readValue, _, err := reader(&output, buffer[:])
			Expect(err).ToNot(HaveOccurred())
			Expect(readValue).To(Equal(value))
		},
		Entry("When using uint16", encoding.EntryLengthEncodingUint16, uint64(1)),
		Entry("When using uint32", encoding.EntryLengthEncodingUint32, uint64(1)),
		Entry("When using uint64", encoding.EntryLengthEncodingUint64, uint64(1)),
		Entry("When using uvarint low", encoding.EntryLengthEncodingUvarint, uint64(1)),
		Entry("When using uvarint MaxUint8", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint8)),
		Entry("When using uvarint MaxUint16", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint16)),
		Entry("When using uvarint MaxUint32", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint32)),
		Entry("When using uvarint MaxUint64", encoding.EntryLengthEncodingUvarint, uint64(math.MaxUint64)),
	)
})

func BenchmarkEntryLengthWriter(b *testing.B) {
	var buffer [encoding.MaxLengthBufferLen]byte
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		writer, err := encoding.GetEntryLengthWriter(entryLengthEncoding)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(entryLengthEncoding.String(), func(b *testing.B) {
			for b.Loop() {
				if err := writer(io.Discard, buffer[:], 1024); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEntryLengthReader(b *testing.B) {
	var buffer [encoding.MaxLengthBufferLen]byte
	for _, entryLengthEncoding := range encoding.EntryLengthEncodings {
		writer, err := encoding.GetEntryLengthWriter(entryLengthEncoding)
		if err != nil {
			b.Fatal(err)
		}
		reader, err := encoding.GetEntryLengthReader(entryLengthEncoding)
		if err != nil {
			b.Fatal(err)
		}
		var output bytes.Buffer
		if err := writer(&output, buffer[:], 1024); err != nil {
			b.Fatal(err)
		}
		outputBackup := output
		// We do the interface conversion outside the loop to not have the allocation of the interface conversion
		// skew the measurements.
		var outputAsIoReader io.Reader = &output
		b.Run(entryLengthEncoding.String(), func(b *testing.B) {
			for b.Loop() {
				output = outputBackup
				if _, _, err := reader(outputAsIoReader, buffer[:]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
