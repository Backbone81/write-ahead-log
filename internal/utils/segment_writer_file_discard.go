package utils

// SegmentWriterFileDiscard provides a stub for a segment file which discards all data. It allows us to run large scale
// benchmarks without filling up the disk or memory.
type SegmentWriterFileDiscard struct{}

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
