package wal_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("Header", func() {
	It("should write the header", func() {
		var output bytes.Buffer
		var buffer [wal.HeaderSize]byte
		Expect(wal.WriteHeader(&output, buffer[:], wal.DefaultHeader)).To(Succeed())
		Expect(output.Len()).To(Equal(wal.HeaderSize))
	})

	It("should read the header", func() {
		var output bytes.Buffer
		var buffer [wal.HeaderSize]byte
		Expect(wal.WriteHeader(&output, buffer[:], wal.DefaultHeader)).To(Succeed())

		gotHeader, err := wal.ReadHeader(&output, buffer[:])
		Expect(err).ToNot(HaveOccurred())

		Expect(gotHeader).To(Equal(wal.DefaultHeader))
	})

	It("should fail reading the header from an empty buffer", func() {
		var input bytes.Buffer
		var buffer [wal.HeaderSize]byte
		Expect(wal.ReadHeader(&input, buffer[:])).Error().To(MatchError(io.EOF))
	})

	It("should fail reading the header with wrong magic bytes", func() {
		var output bytes.Buffer
		var buffer [wal.HeaderSize]byte
		Expect(wal.WriteHeader(&output, buffer[:], wal.DefaultHeader)).To(Succeed())

		output.Bytes()[2] = 'X'
		Expect(wal.ReadHeader(&output, buffer[:])).Error().To(MatchError(wal.ErrHeaderInvalidMagicBytes))
	})

	It("should fail reading the header which is too short", func() {
		var output bytes.Buffer
		var buffer [wal.HeaderSize]byte
		Expect(wal.WriteHeader(&output, buffer[:], wal.DefaultHeader)).To(Succeed())

		output.Truncate(output.Len() - 1)
		Expect(wal.ReadHeader(&output, buffer[:])).Error().To(MatchError(io.ErrUnexpectedEOF))
	})
})

func BenchmarkWriteHeader(b *testing.B) {
	var buffer [wal.HeaderSize]byte
	for b.Loop() {
		if err := wal.WriteHeader(io.Discard, buffer[:], wal.DefaultHeader); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadHeader(b *testing.B) {
	var output bytes.Buffer
	var buffer [wal.HeaderSize]byte
	if err := wal.WriteHeader(&output, buffer[:], wal.DefaultHeader); err != nil {
		b.Fatal(err)
	}

	input := SegmentReaderFileLoop{
		Data: output.Bytes(),
	}

	for b.Loop() {
		if _, err := wal.ReadHeader(&input, buffer[:]); err != nil {
			b.Fatal(err)
		}
	}
}
