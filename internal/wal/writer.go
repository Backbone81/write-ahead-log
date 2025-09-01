package wal

import (
	"fmt"
	"path"
	"sync"
)

type Writer struct {
	sync.Mutex

	segmentWriter  *SegmentWriter
	maxSegmentSize int64
}

func (w *Writer) AppendEntry(data []byte) error {
	if err := w.segmentWriter.AppendEntry(data); err != nil {
		return fmt.Errorf("writing entry to segment file: %w", err)
	}

	return nil
}

func (w *Writer) Close() error {
	return w.segmentWriter.Close()
}

func (w *Writer) RolloverIfNeeded(syncPolicy SyncPolicy) error {
	if w.segmentWriter.Offset() < w.maxSegmentSize {
		// We did not yet reach the desired maximum segment size. We can continue with what we have at hand.
		return nil
	}

	return w.Rollover(syncPolicy)
}

func (w *Writer) Rollover(syncPolicy SyncPolicy) error {
	if err := w.segmentWriter.Close(); err != nil {
		return err
	}

	nextSegmentWriter, err := CreateSegment(path.Dir(w.segmentWriter.FilePath()), w.segmentWriter.NextSequenceNumber(), w.maxSegmentSize, syncPolicy)
	if err != nil {
		return err
	}
	w.segmentWriter = nextSegmentWriter
	return nil
}
