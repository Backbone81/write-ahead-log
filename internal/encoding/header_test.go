package encoding_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/write-ahead-log/internal/encoding"
	"github.com/backbone81/write-ahead-log/internal/utils"
)

var _ = Describe("Header", func() {
	It("should write the header", func() {
		var output bytes.Buffer
		var buffer [encoding.HeaderSize]byte
		Expect(encoding.WriteHeader(&output, buffer[:], encoding.DefaultHeader)).To(Succeed())
		Expect(output.Len()).To(Equal(encoding.HeaderSize))
	})

	It("should read the header", func() {
		var output bytes.Buffer
		var buffer [encoding.HeaderSize]byte
		Expect(encoding.WriteHeader(&output, buffer[:], encoding.DefaultHeader)).To(Succeed())

		gotHeader, err := encoding.ReadHeader(&output, buffer[:])
		Expect(err).ToNot(HaveOccurred())

		Expect(gotHeader).To(Equal(encoding.DefaultHeader))
	})

	It("should fail reading the header from an empty buffer", func() {
		var input bytes.Buffer
		var buffer [encoding.HeaderSize]byte
		Expect(encoding.ReadHeader(&input, buffer[:])).Error().To(MatchError(io.EOF))
	})

	It("should fail reading the header with wrong magic bytes", func() {
		var output bytes.Buffer
		var buffer [encoding.HeaderSize]byte
		Expect(encoding.WriteHeader(&output, buffer[:], encoding.DefaultHeader)).To(Succeed())

		output.Bytes()[2] = 'X'
		Expect(encoding.ReadHeader(&output, buffer[:])).Error().To(MatchError(encoding.ErrHeaderInvalidMagicBytes))
	})

	It("should fail reading the header which is too short", func() {
		var output bytes.Buffer
		var buffer [encoding.HeaderSize]byte
		Expect(encoding.WriteHeader(&output, buffer[:], encoding.DefaultHeader)).To(Succeed())

		output.Truncate(output.Len() - 1)
		Expect(encoding.ReadHeader(&output, buffer[:])).Error().To(MatchError(io.ErrUnexpectedEOF))
	})
})

func BenchmarkWriteHeader(b *testing.B) {
	var buffer [encoding.HeaderSize]byte
	for b.Loop() {
		if err := encoding.WriteHeader(io.Discard, buffer[:], encoding.DefaultHeader); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadHeader(b *testing.B) {
	var output bytes.Buffer
	var buffer [encoding.HeaderSize]byte
	if err := encoding.WriteHeader(&output, buffer[:], encoding.DefaultHeader); err != nil {
		b.Fatal(err)
	}

	input := utils.SegmentReaderFileLoop{
		Data: output.Bytes(),
	}

	for b.Loop() {
		if _, err := encoding.ReadHeader(&input, buffer[:]); err != nil {
			b.Fatal(err)
		}
	}
}
