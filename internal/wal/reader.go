package wal

import (
	"errors"
	"fmt"
	"io"
	"time"

	"write-ahead-log/internal/utils"
)

// Reader provides functionality to read the write-ahead-log. It abstracts away the fact that the write-ahead log is
// split into multiple segments.
//
// Instances of this struct are NOT safe for concurrent use. Either use it on a single Go routine or provide your own
// external synchronization.
type Reader struct {
	noCopy utils.NoCopy

	// The currently active segment reader where we are reading entries from.
	segmentReader *SegmentReader

	// The error for the last operation. If this is nil, the value of the segment reader can be used. We need to keep
	// our own err around, because there are errors which can occur when opening segment files. Therefore, we can not
	// rely on the segment reader err only.
	err error
}

// NewReader creates a new Reader starting at the given sequence number. It will find the segment the sequence number
// belongs to and read all entries up until the requested sequence number.
func NewReader(directory string, sequenceNumber uint64) (*Reader, error) {
	// Identify which segment contains the requested sequence number. The segment itself is the first sequence number
	// in the segment.
	segment, err := SegmentFromSequenceNumber(directory, sequenceNumber)
	if err != nil {
		return nil, err
	}

	// Create a segment reader for the given segment and make sure that the segment file name actually matches to the
	// first sequence number as documented in the segment header.
	segmentReader, err := OpenSegment(directory, segment)
	if err != nil {
		return nil, err
	}

	// Move the WAL reader forward until we have reached the desired sequence number.
	newReader := Reader{
		segmentReader: segmentReader,
	}
	for newReader.NextSequenceNumber() < sequenceNumber && newReader.Next() {
		// Skip entry until we have reached our target sequence number.
	}
	if newReader.Err() != nil {
		// We abort here if we are unable to reach the requested location.
		return nil, newReader.Err()
	}
	if newReader.NextSequenceNumber() != sequenceNumber {
		// This should never happen, when we did not get any error from Next(), but we still double check.
		return nil, fmt.Errorf("expected to reach sequence number %d but instead reached %d", sequenceNumber, newReader.NextSequenceNumber())
	}

	return &newReader, nil
}

// FilePath returns the file path of the file this reader is reading from.
func (r *Reader) FilePath() string {
	return r.segmentReader.FilePath()
}

// Header returns the segment file header.
func (r *Reader) Header() Header {
	return r.segmentReader.Header()
}

// Offset returns the offset in bytes from the start of the file.
func (r *Reader) Offset() int64 {
	return r.segmentReader.Offset()
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (r *Reader) NextSequenceNumber() uint64 {
	return r.segmentReader.NextSequenceNumber()
}

// Next reports if an entry has been successfully read. When it returns true, Err() returns nil and Value() contains
// valid data. When it returns false, Err() returns an error. Value() contains invalid data in that situation.
func (r *Reader) Next() bool {
	// Forward to our active segment reader first.
	next := r.segmentReader.Next()
	r.err = r.segmentReader.Err()

	if next {
		// As we successfully read an entry from the segment, there is nothing else to do.
		return true
	}

	if !errors.Is(r.err, io.EOF) {
		// Any error other than end of file results in an early exit. In case of end of file, we want to replace
		// the current segment reader with the next segment reader.
		return false
	}

	nextSegmentReader, err := OpenSegment(SegmentFileName(r.segmentReader.NextSequenceNumber()), r.segmentReader.NextSequenceNumber())
	if err != nil {
		// We keep the old error in r.err because this wil still signal that no entry could be read.
		return false
	}

	// We are ready to move on to the next segment reader, so close our active one.
	if err := r.segmentReader.Close(); err != nil {
		_ = nextSegmentReader.Close()
		r.err = fmt.Errorf("closing the segment reader: %w", err)
		return false
	}

	// Replace our current segment reader with the next segment reader and call recursively into Next() to deal with
	// potential errors with the next segment reader there.
	r.segmentReader = nextSegmentReader
	return r.Next()
}

// Value returns the last entry read from the segment file. The values are only valid after the first call to Next()
// and while Err() is nil.
func (r *Reader) Value() SegmentReaderValue {
	return r.segmentReader.Value()
}

// Err returns the error for the last call to Next().
func (r *Reader) Err() error {
	return r.err
}

// ToWriter returns a writer to append entries to the write-ahead log. This is the only way to create a writer, because
// we can only know if we have reached the end of the segment, when we read all elements from it. Creating a writer
// will fail, when not all entries were read.
// The reader must not be used any more after a call to this function.
func (r *Reader) ToWriter(options ...WriterOption) (*Writer, error) {
	newWriter := Writer{
		preAllocationSize:   DefaultPreAllocationSize,
		maxSegmentSize:      DefaultPreAllocationSize,
		entryLengthEncoding: r.segmentReader.Header().EntryLengthEncoding,
		entryChecksumType:   r.segmentReader.Header().EntryChecksumType,
		rolloverCallback:    DefaultRolloverCallback,
	}
	newWriter.syncPolicy = NewSyncPolicyGrouped(10*time.Millisecond, &newWriter.Mutex)
	for _, option := range options {
		option(&newWriter)
	}

	newSegmentWriter, err := r.segmentReader.ToWriter(newWriter.syncPolicy)
	if err != nil {
		return nil, err
	}

	newWriter.segmentWriter = newSegmentWriter
	return &newWriter, nil
}

// Close closes the underlying reader.
func (r *Reader) Close() error {
	return r.segmentReader.Close()
}
