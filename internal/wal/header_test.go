package wal_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

var _ = Describe("Header", func() {
	It("should write the header", func() {
		var buffer bytes.Buffer
		header := wal.Header{
			Magic:               wal.Magic,
			Version:             1,
			FirstSequenceNumber: 7,
		}
		Expect(header.Write(&buffer)).To(Succeed())
		Expect(buffer.Len()).To(Equal(wal.HeaderSize))
	})

	It("should read the header", func() {
		var buffer bytes.Buffer
		wantHeader := wal.Header{
			Magic:               wal.Magic,
			Version:             1,
			FirstSequenceNumber: 7,
		}
		Expect(wantHeader.Write(&buffer)).To(Succeed())

		var gotHeader wal.Header
		Expect(gotHeader.Read(&buffer)).To(Succeed())

		Expect(gotHeader).To(Equal(wantHeader))
	})

	It("should fail reading the header from an empty buffer", func() {
		var buffer bytes.Buffer
		var header wal.Header
		Expect(header.Read(&buffer)).ToNot(Succeed())
	})

	It("should fail reading the header with wrong magic bytes", func() {
		var buffer bytes.Buffer
		header := wal.Header{
			Magic:               wal.Magic,
			Version:             1,
			FirstSequenceNumber: 7,
		}
		Expect(header.Write(&buffer)).To(Succeed())

		buffer.Bytes()[2] = 'X'
		Expect(header.Read(&buffer)).ToNot(Succeed())
	})

	It("should fail reading the header which is too short", func() {
		var buffer bytes.Buffer
		header := wal.Header{
			Magic:               wal.Magic,
			Version:             1,
			FirstSequenceNumber: 7,
		}
		Expect(header.Write(&buffer)).To(Succeed())

		buffer.Truncate(buffer.Len() - 1)
		Expect(header.Read(&buffer)).ToNot(Succeed())
	})
})

func BenchmarkHeader_Write(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	header := wal.Header{
		Magic:               wal.Magic,
		Version:             1,
		FirstSequenceNumber: 0,
	}
	if err := header.Validate(); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for range b.N {
		if err := header.Write(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHeader_Read(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	header := wal.Header{
		Magic:               wal.Magic,
		Version:             1,
		FirstSequenceNumber: 0,
	}
	for range b.N {
		if err := header.Write(buffer); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()

	for range b.N {
		if err := header.Read(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
