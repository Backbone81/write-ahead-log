package wal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"write-ahead-log/internal/utils"
)

type SegmentWriter struct {
	noCopy utils.NoCopy

	file *os.File

	mutex              sync.Mutex
	nextSequenceNumber uint64
	currOffset         int64
	syncPolicy         SyncPolicy
}

func CreateSegment(directory string, firstSequenceNumber uint64, maxSegmentSize int64) (*SegmentWriter, error) {
	newSegmentFileName := segmentFileName(firstSequenceNumber) + ".new"
	newSegmentFilePath := path.Join(directory, newSegmentFileName)
	if err := os.Remove(newSegmentFilePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("removing the segment file %q: %w", newSegmentFilePath, err)
	}

	segmentFile, err := os.OpenFile(newSegmentFilePath, os.O_RDWR|os.O_CREATE, 0o664) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("creating the segment file %q: %w", newSegmentFilePath, err)
	}

	if err := segmentFile.Truncate(maxSegmentSize); err != nil {
		return nil, fmt.Errorf("pre-allocating the segment file %q: %w", newSegmentFilePath, err)
	}

	walHeader := Header{
		Magic:               Magic,
		Version:             1,
		FirstSequenceNumber: firstSequenceNumber,
	}
	if err := walHeader.Write(segmentFile); err != nil {
		return nil, fmt.Errorf("writing header to segment file %q: %w", newSegmentFilePath, err)
	}
	if err := segmentFile.Sync(); err != nil {
		return nil, fmt.Errorf("flushing the segment file %q: %w", newSegmentFilePath, err)
	}

	segmentFilePath := path.Join(directory, segmentFileName(firstSequenceNumber))
	if err := os.Rename(newSegmentFilePath, segmentFilePath); err != nil {
		return nil, fmt.Errorf("renaming the segment file from %q to %q: %w", newSegmentFilePath, segmentFilePath, err)
	}
	return &SegmentWriter{
		file: segmentFile,
	}, nil
}

func newSegmentWriterFromFile(segmentFile *os.File, nextSequenceNumber uint64) (*SegmentWriter, error) {
	currOffset, err := segmentFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading file position: %w", err)
	}

	return &SegmentWriter{
		file:               segmentFile,
		nextSequenceNumber: nextSequenceNumber,
		currOffset:         currOffset,
	}, nil
}

func (w *SegmentWriter) Close() error {
	syncErr := w.syncPolicy.Close()
	closeErr := w.file.Close()
	return errors.Join(syncErr, closeErr)
}

// GetOffset returns the current offset in bytes from the start of the segment file.
func (w *SegmentWriter) GetOffset() int64 {
	return w.currOffset
}

func (w *SegmentWriter) AppendEntry(data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if err := WriteEntry(w.file, data); err != nil {
		return fmt.Errorf("writing entry to segment file: %w", err)
	}
	currSequenceNumber := w.nextSequenceNumber
	w.nextSequenceNumber++

	if err := w.syncPolicy.EntryAppended(currSequenceNumber); err != nil {
		return fmt.Errorf("flushing entry to segment file: %w", err)
	}
	return nil
}
