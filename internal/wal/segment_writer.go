package wal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"write-ahead-log/internal/utils"
)

// SegmentWriter provides functionality for writing to a single segment file.
//
// Instances of SegmentWriter are NOT safe to use concurrently. You need to provide external synchronization.
type SegmentWriter struct {
	noCopy utils.NoCopy

	// The file the writer is writing data to.
	file *os.File

	// This buffer is used to combine multiple individual file write commands into a single one to improve performance.
	buffer *bytes.Buffer

	// The sequence number the next entry will receive.
	nextSequenceNumber uint64

	// The current offset in bytes from the start of the file.
	offset int64

	// The policy describing how data is flushed to disk.
	syncPolicy SyncPolicy

	entryLengthWriter   EntryLengthWriter
	entryChecksumWriter EntryChecksumWriter
}

// CreateSegment creates a new segment file in the given directory. It will create the new file with the file extension
// ".new" appended to the file name and rename it after the header has been written to. This ensures that the new
// segment file is only visible in the directory when the header was correctly written and flushed to stable storage.
//
// directory is the directory all segment files are located in.
// firstSequenceNumber is used for deriving the file name and for storing it in the segment header.
// segmentSize is the size of the file which is pre-allocated.
// syncPolicy describes how changes are flushed to stable storage.
func CreateSegment(directory string, firstSequenceNumber uint64, segmentSize int64, syncPolicy SyncPolicy) (*SegmentWriter, error) {
	// Remove any temporary segment file which might be there from an earlier failure.
	newSegmentFileName := segmentFileName(firstSequenceNumber) + ".new"
	newSegmentFilePath := path.Join(directory, newSegmentFileName)
	if err := os.Remove(newSegmentFilePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("removing the segment file %q: %w", newSegmentFilePath, err)
	}

	// Create the temporary segment file and pre-allocate its size.
	segmentFile, err := os.OpenFile(newSegmentFilePath, os.O_RDWR|os.O_CREATE, 0o664) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("creating the segment file %q: %w", newSegmentFilePath, err)
	}

	if err := segmentFile.Truncate(segmentSize); err != nil {
		return nil, fmt.Errorf("pre-allocating the segment file %q: %w", newSegmentFilePath, err)
	}

	// Write the header to the segment file and flush the content to stable storage.
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

	// Rename the temporary segment file to the final one.
	segmentFilePath := path.Join(directory, segmentFileName(firstSequenceNumber))
	if err := os.Rename(newSegmentFilePath, segmentFilePath); err != nil {
		return nil, fmt.Errorf("renaming the segment file from %q to %q: %w", newSegmentFilePath, segmentFilePath, err)
	}
	return newSegmentWriterFromFile(segmentFile, firstSequenceNumber, syncPolicy)
}

// newSegmentWriterFromFile creates a SegmentWriter from a file which is already open.
func newSegmentWriterFromFile(segmentFile *os.File, nextSequenceNumber uint64, syncPolicy SyncPolicy) (*SegmentWriter, error) {
	offset, err := segmentFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading file position: %w", err)
	}

	return &SegmentWriter{
		file:               segmentFile,
		nextSequenceNumber: nextSequenceNumber,
		offset:             offset,
		syncPolicy:         syncPolicy,
		buffer:             bytes.NewBuffer(make([]byte, 0, 1024)),
	}, nil
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (w *SegmentWriter) NextSequenceNumber() uint64 {
	return w.nextSequenceNumber
}

// Offset returns the offset in bytes from the start of the file.
func (w *SegmentWriter) Offset() int64 {
	return w.offset
}

// AppendEntry adds the given entry to the segment.
func (w *SegmentWriter) AppendEntry(data []byte) error {
	w.buffer.Reset()
	n, err := WriteEntry(w.buffer, data)
	if err != nil {
		return fmt.Errorf("writing entry to buffer: %w", err)
	}
	if _, err := w.file.Write(w.buffer.Bytes()); err != nil {
		return fmt.Errorf("writing entry to segment file: %w", err)
	}
	sequenceNumber := w.nextSequenceNumber
	w.nextSequenceNumber++
	w.offset += int64(n)

	if err := w.syncPolicy.EntryAppended(sequenceNumber); err != nil {
		return fmt.Errorf("flushing entry to segment file: %w", err)
	}
	return nil
}

// Close flushes all pending changes to disk and closes the file.
func (w *SegmentWriter) Close() error {
	syncErr := w.syncPolicy.Close()
	closeErr := w.file.Close()
	return errors.Join(syncErr, closeErr)
}
