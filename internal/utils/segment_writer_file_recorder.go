package utils

import (
	"bytes"
)

// SegmentWriterFileRecorder provides a stub for a segment file which records what is written to it in memory. It allows
// us to use a SegmentWriter to prepare a buffer which can then be used by SegmentReaderFileLoop to serve read requests.
type SegmentWriterFileRecorder struct {
	bytes.Buffer
}

func (s *SegmentWriterFileRecorder) Close() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Sync() error {
	return nil
}

func (s *SegmentWriterFileRecorder) Name() string {
	return "in-memory-recorder"
}

func (s *SegmentWriterFileRecorder) Truncate(offset int64) error {
	return nil
}
