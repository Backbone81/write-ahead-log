package wal_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"write-ahead-log/internal/wal"
)

func TestWal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WAL Suite")
}

// SegmentWriterFileDiscard provides a stub for a segment file which discards all data. It allows us to run large scale
// benchmarks without filling up the disk or memory.
type SegmentWriterFileDiscard struct{}

// SegmentWriterFileDiscard implements SegmentWriterFile.
var _ wal.SegmentWriterFile = (*SegmentWriterFileDiscard)(nil)

func (s *SegmentWriterFileDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (s *SegmentWriterFileDiscard) Close() error {
	return nil
}

func (s *SegmentWriterFileDiscard) Sync() error {
	return nil
}

func (s *SegmentWriterFileDiscard) Name() string {
	return "in-memory-discard"
}

// SegmentReaderFileLoop provides a stub for the segment file which returns the same data over and over again in an
// endless loop. It allows us to run large scale benchmarks without having to provide an actual big file on disk or
// memory.
type SegmentReaderFileLoop struct {
	Data   []byte
	Offset int
}

// SegmentReaderFileLoop implements SegmentReaderFile.
var _ wal.SegmentReaderFile = (*SegmentReaderFileLoop)(nil)

func (s *SegmentReaderFileLoop) Read(p []byte) (int, error) {
	copyBytes := min(len(p), len(s.Data)-s.Offset)
	copy(p, s.Data[s.Offset:s.Offset+copyBytes])
	s.Offset += copyBytes
	if s.Offset >= len(s.Data) {
		s.Offset = 0
	}
	return copyBytes, nil
}

func (s *SegmentReaderFileLoop) Close() error {
	return nil
}

func (s *SegmentReaderFileLoop) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (s *SegmentReaderFileLoop) Name() string {
	return "in-memory-loop"
}

// SegmentWriterFileRecorder provides a stub for a segment file which records what is written to it in memory. It allows
// us to use a SegmentWriter to prepare a buffer which can then be used by SegmentReaderFileLoop to serve read requests.
type SegmentWriterFileRecorder struct {
	bytes.Buffer
}

// SegmentWriterFileRecorder implements SegmentWriterFile.
var _ wal.SegmentWriterFile = (*SegmentWriterFileRecorder)(nil)

func (s *SegmentWriterFileRecorder) Close() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Sync() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Name() string {
	return "in-memory-recorder"
}
