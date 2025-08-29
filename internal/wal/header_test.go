package wal_test

import (
	"io"
	"testing"
	"write-ahead-log/internal/wal"
)

func BenchmarkHeader_Write(b *testing.B) {
	header := wal.Header{
		Magic:               wal.Magic,
		Version:             1,
		FirstSequenceNumber: 0,
	}
	if err := header.Validate(); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err := header.Write(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
