package wal_test

import (
	"io"
	"testing"
	"write-ahead-log/internal/wal"
)

func BenchmarkWriteEntryNil(b *testing.B) {
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEntry0K(b *testing.B) {
	var data [0 * 1024]byte
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, data[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEntry1K(b *testing.B) {
	var data [1 * 1024]byte
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, data[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEntry2K(b *testing.B) {
	var data [2 * 1024]byte
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, data[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEntry4K(b *testing.B) {
	var data [4 * 1024]byte
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, data[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEntry8K(b *testing.B) {
	var data [8 * 1024]byte
	for b.Loop() {
		if err := wal.WriteEntry(io.Discard, data[:]); err != nil {
			b.Fatal(err)
		}
	}
}
