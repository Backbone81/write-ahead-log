package wal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"write-ahead-log/internal/utils"
)

// SegmentReader provides functionality for reading a segment file of the write-ahead-log.
//
// It is not thread safe and should only be used in a single go routine. Otherwise, external synchronization must be
// provided.
type SegmentReader struct {
	noCopy utils.NoCopy

	// The segment file to read from.
	file *os.File

	// The file header as read from the start of the segment file.
	header Header

	// The total size of the file in bytes. This is used together with currOffset to calculate the available data until
	// the end of file. This helps with avoiding large memory allocations with malformed files.
	fileSize int64

	// The current offset from the start of the file in bytes. This is used together with fileSize to calculate the
	// available data until the end of the file, and to reset to a former offset after a failed read of an entry.
	currOffset int64

	// The next sequence number is used to keep track of the sequence number as we are reading entries from the segment.
	nextSequenceNumber uint64

	// The value the segment reader returns. Only contains useful data if err is nil.
	value SegmentReaderValue

	// The error for the last operation. If this is nil, the content of value can be used.
	err error
}

// SegmentReaderValue is the value returned by the SegmentReader.
type SegmentReaderValue struct {
	// The sequence number of the entry.
	SequenceNumber uint64

	// The data of the entry.
	Data []byte
}

// NewSegmentReader creates a new segment reader for the file path given as parameter.
//
// To avoid resources leaking, the returned SegmentReader needs to be closed by calling Close().
// Returns an error if the file cannot be opened, read from or the header is malformed.
func NewSegmentReader(directory string, firstSequenceNumber uint64) (*SegmentReader, error) {
	segmentFilePath := path.Join(directory, segmentFileName(firstSequenceNumber))
	segmentReader, err := newSegmentReader(segmentFilePath, firstSequenceNumber)
	if err != nil {
		return nil, fmt.Errorf("segment file %q: %w", segmentFilePath, err)
	}
	return segmentReader, nil
}

func newSegmentReader(segmentFilePath string, firstSequenceNumber uint64) (*SegmentReader, error) {
	segmentFile, err := os.OpenFile(segmentFilePath, os.O_RDWR, 0) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	segmentReader, err := newSegmentReaderFromFile(segmentFile, firstSequenceNumber)
	if err != nil {
		if closeErr := segmentFile.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return segmentReader, nil
}

func newSegmentReaderFromFile(segmentFile *os.File, firstSequenceNumber uint64) (*SegmentReader, error) {
	var header Header
	if err := header.Read(segmentFile); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if err := header.Validate(); err != nil {
		return nil, fmt.Errorf("validating header: %w", err)
	}
	if header.FirstSequenceNumber != firstSequenceNumber {
		return nil, fmt.Errorf("expected first sequence number to be %d but got %d", firstSequenceNumber, header.FirstSequenceNumber)
	}

	fileInfo, err := segmentFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading file size: %w", err)
	}

	currOffset, err := segmentFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading file position: %w", err)
	}

	return &SegmentReader{
		file:               segmentFile,
		header:             header,
		fileSize:           fileInfo.Size(),
		currOffset:         currOffset,
		nextSequenceNumber: firstSequenceNumber,
		value: SegmentReaderValue{
			// Pre-allocate the data slice to reduce the number of allocations.
			Data: make([]byte, 0, 4*1024),
		},
		err: nil,
	}, nil
}

// Close closes the file the SegmentReader is reading from.
func (r *SegmentReader) Close() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	return nil
}

func (r *SegmentReader) Header() Header {
	return r.header
}

// Offset returns the current offset in bytes from the start of the segment file.
func (r *SegmentReader) Offset() int64 {
	return r.currOffset
}

func (r *SegmentReader) NextSequenceNumber() uint64 {
	return r.nextSequenceNumber
}

// Next reports if an entry has been successfully read. When it returns true, Err() returns nil and Value() contains
// valid data. When it returns false, Err() might be nil if the reader has reached the end of the file, or it might
// return an error. Value() contains invalid data in that situation.
func (r *SegmentReader) Next() bool {
	var n int
	n, r.value.Data, r.err = ReadEntry(r.file, r.value.Data, r.fileSize-r.currOffset)
	if r.err != nil {
		// In case of an error when reading the next entry, we move the file position back to where we were before.
		// Otherwise, we could not reliably continue writing to a segment file which has not yet reached the desired
		// maximum size.
		if _, err := r.file.Seek(r.currOffset, io.SeekStart); err != nil {
			r.err = errors.Join(r.err, err)
		}
		return false
	}
	r.currOffset += int64(n)
	r.value.SequenceNumber = r.nextSequenceNumber
	r.nextSequenceNumber++
	return true
}

// Value returns the last entry read from the segment file. The values are only valid after the first call to Next()
// and while Err() is nil.
func (r *SegmentReader) Value() SegmentReaderValue {
	return r.value
}

// Err returns the error for the last call to Next().
func (r *SegmentReader) Err() error {
	return r.err
}

// ToWriter returns a SegmentWriter to append to the open segment file. You must have read all entries of the segment
// before you call this method. Otherwise, it will fail. After a call to ToWriter(), you cannot use the SegmentReader
// anymore.
func (r *SegmentReader) ToWriter(syncPolicy SyncPolicy) (*SegmentWriter, error) {
	if !errors.Is(r.err, ErrEntryNone) {
		return nil, errors.New("segment needs to be read until the last entry is reached")
	}

	segmentWriter, err := NewSegmentWriter(r.file.Name(), r.file, r.header, r.currOffset, r.nextSequenceNumber, syncPolicy)
	if err != nil {
		return nil, err
	}

	// Make sure this reader is not used for anything else afterward.
	*r = SegmentReader{}
	return segmentWriter, nil
}
