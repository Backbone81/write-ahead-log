package utils

// SegmentReaderFileLoop provides a stub for the segment file which returns the same data over and over again in an
// endless loop. It allows us to run large scale benchmarks without having to provide an actual big file on disk or
// memory.
type SegmentReaderFileLoop struct {
	Data   []byte
	Offset int
}

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
