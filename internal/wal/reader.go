package wal

import (
	"errors"
	"fmt"
	"io"

	"write-ahead-log/internal/utils"
)

// Reader provides functionality to read the write-ahead-log. It abstracts away the fact that the write-ahead-log is
// split into multiple segments.
//
// Instances of this struct are NOT safe for concurrent use. Either use it on a single Go routine or provide your own
// external synchronization.
type Reader struct {
	noCopy utils.NoCopy

	// The currently active segment reader where we are reading entries from.
	segmentReader *SegmentReader

	// The sequence number of the entry we read next. We keep track of it to know what the filename of the next segment
	// file is when we hit the end of the file of the current segment.
	nextSequenceNumber uint64

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
	segmentReader, err := OpenSegment(segmentFileName(segment), segment)
	if err != nil {
		return nil, err
	}

	// Move the WAL reader forward until we have reached the desired sequence number.
	newReader := Reader{
		segmentReader:      segmentReader,
		nextSequenceNumber: sequenceNumber,
	}
	for newReader.nextSequenceNumber < sequenceNumber && newReader.Next() {
		// Skip entry until we have reached our target sequence number.
	}
	if newReader.Err() != nil {
		// We abort here if we are unable to reach the requested location.
		return nil, newReader.Err()
	}
	if newReader.nextSequenceNumber != sequenceNumber {
		// This should never happen, when we did not get any error from Next(), but we still double check.
		return nil, fmt.Errorf("expected to reach sequence number %d but instead reached %d", sequenceNumber, newReader.nextSequenceNumber)
	}

	return &newReader, nil
}

func (r *Reader) Close() error {
	return r.segmentReader.Close()
}

// Next reports if an entry has been successfully read. When it returns true, Err() returns nil and Value() contains
// valid data. When it returns false, Err() might be nil if the reader has reached the end of the file, or it might
// return an error. Value() contains invalid data in that situation.
func (r *Reader) Next() bool {
	// Forward to our active segment reader first.
	next := r.segmentReader.Next()
	r.err = r.segmentReader.Err()

	if next {
		// The segment reader successfully read an entry. We need to keep track of the next sequence number.
		r.nextSequenceNumber++
		return true
	}

	if !errors.Is(r.err, io.EOF) {
		// Any error other than end of file results in an early exit. In case of end of file, we want to replace
		// the current segment reader with the next segment reader.
		return false
	}

	// We are ready to move on to the next segment reader, so close our active one.
	if err := r.segmentReader.Close(); err != nil {
		r.err = fmt.Errorf("closing the segment reader: %w", err)
		return false
	}

	nextSegmentReader, err := OpenSegment(segmentFileName(r.nextSequenceNumber), r.nextSequenceNumber)
	if err != nil {
		r.err = err
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

func (r *Reader) ToWriter(maxSegmentSize int64, syncPolicyType SyncPolicyType) (*Writer, error) {
	newSegmentWriter, err := r.segmentReader.ToWriter(syncPolicyType)
	if err != nil {
		return nil, err
	}

	newWriter := Writer{
		segmentWriter: newSegmentWriter,
	}
	if err := newWriter.RolloverIfNeeded(syncPolicyType); err != nil {
		return nil, err
	}
	return &newWriter, nil
}
