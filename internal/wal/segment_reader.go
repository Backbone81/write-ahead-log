package wal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"write-ahead-log/internal/utils"
)

var ErrEntryNone = errors.New("this is no WAL entry")

// SegmentReaderFile is an interface which needs to be implemented by the file to read from.
type SegmentReaderFile interface {
	io.ReadCloser
	io.Seeker
	Name() string
}

// SegmentReader provides functionality for reading a segment file of the write-ahead-log.
//
// It is not thread safe and should only be used in a single go routine. Otherwise, external synchronization must be
// provided.
type SegmentReader struct {
	noCopy utils.NoCopy

	// The segment file to read from.
	file SegmentReaderFile

	// The file header as read from the start of the segment file.
	header Header

	// The current offset from the start of the file in bytes. This is used together with fileSize to calculate the
	// available data until the end of the file, and to reset to a former offset after a failed read of an entry.
	offset int64

	// The next sequence number is used to keep track of the sequence number as we are reading entries from the segment.
	nextSequenceNumber uint64

	// The reader to decode the length of an entry.
	entryLengthReader EntryLengthReader

	// The reader to calculate and read the checksum.
	entryChecksumReader EntryChecksumReader

	// The buffer to hold the entry data.
	data []byte

	// The total size of the file in bytes. This is used together with offset to calculate the available data until
	// the end of file. This helps with avoiding large memory allocations with malformed files.
	fileSize int64

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

// OpenSegment creates a new segment reader for the file path given as parameter.
//
// To avoid resources leaking, the returned SegmentReader needs to be closed by calling Close().
// Returns an error if the file cannot be opened, read from or the header is malformed.
func OpenSegment(directory string, firstSequenceNumber uint64) (*SegmentReader, error) {
	segmentFilePath := path.Join(directory, segmentFileName(firstSequenceNumber))
	segmentReader, err := openSegment(segmentFilePath, firstSequenceNumber)
	if err != nil {
		return nil, fmt.Errorf("segment file %q: %w", segmentFilePath, err)
	}
	return segmentReader, nil
}

func openSegment(segmentFilePath string, firstSequenceNumber uint64) (*SegmentReader, error) {
	segmentFile, err := os.OpenFile(segmentFilePath, os.O_RDWR, 0) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	var segmentHeader Header
	if err := segmentHeader.Read(segmentFile); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if err := segmentHeader.Validate(); err != nil {
		return nil, fmt.Errorf("validating header: %w", err)
	}
	if segmentHeader.FirstSequenceNumber != firstSequenceNumber {
		return nil, fmt.Errorf("expected first sequence number to be %d but got %d", firstSequenceNumber, segmentHeader.FirstSequenceNumber)
	}

	fileInfo, err := segmentFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading file size: %w", err)
	}

	currOffset, err := segmentFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading file position: %w", err)
	}

	segmentReader, err := NewSegmentReader(segmentFile, segmentHeader, fileInfo.Size(), currOffset, firstSequenceNumber)
	if err != nil {
		if closeErr := segmentFile.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return segmentReader, nil
}

func NewSegmentReader(segmentFile SegmentReaderFile, segmentHeader Header, fileSize int64, offset int64, nextSequenceNumber uint64) (*SegmentReader, error) {
	entryLengthReader, err := GetEntryLengthReader(segmentHeader.EntryLengthEncoding)
	if err != nil {
		return nil, err
	}

	entryChecksumReader, err := GetEntryChecksumReader(segmentHeader.EntryChecksumType)
	if err != nil {
		return nil, err
	}

	return &SegmentReader{
		file:                segmentFile,
		header:              segmentHeader,
		offset:              offset,
		nextSequenceNumber:  nextSequenceNumber,
		entryLengthReader:   entryLengthReader,
		entryChecksumReader: entryChecksumReader,
		data:                make([]byte, 4*1024), // Pre-allocate the data slice to reduce the number of allocations.
		fileSize:            fileSize,
	}, nil
}

// FilePath returns the file path of the file this reader is reading from.
func (r *SegmentReader) FilePath() string {
	return r.file.Name()
}

// Header returns the segment file header.
func (r *SegmentReader) Header() Header {
	return r.header
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (r *SegmentReader) NextSequenceNumber() uint64 {
	return r.nextSequenceNumber
}

// Offset returns the offset in bytes from the start of the file.
func (r *SegmentReader) Offset() int64 {
	return r.offset
}

// Next reports if an entry has been successfully read. When it returns true, Err() returns nil and Value() contains
// valid data. When it returns false, Err() might be nil if the reader has reached the end of the file, or it might
// return an error. Value() contains invalid data in that situation.
func (r *SegmentReader) Next() bool {
	if r.err = r.next(); r.err != nil {
		r.err = errors.Join(ErrEntryNone, r.err)

		// In case of an error when reading the next entry, we move the file position back to where we were before.
		// Otherwise, we could not reliably continue writing to a segment file which has not yet reached the desired
		// maximum size.
		if _, err := r.file.Seek(r.offset, io.SeekStart); err != nil {
			r.err = errors.Join(r.err, err)
		}
		return false
	}
	return true
}

func (r *SegmentReader) next() error {
	// Read the length of the entry.
	// We use the data slice as scratch space for converting bytes to integers. We assume that the data slice can always
	// hold at least the maximum length encoding. This is true for a pre-allocated data slice.
	length, lengthBytes, err := r.entryLengthReader(r.file, r.data[:MaxLengthBufferLen])
	if err != nil {
		return err
	}

	remainingBytes := r.fileSize - r.offset
	if remainingBytes < int64(length) { //nolint:gosec // chances are low that length will overflow
		return errors.New("the WAL entry data exceeds the maximum possible size")
	}

	// Read the data part of the entry.
	// As we are using the data slice as scratch space as well, we need to make sure that we not only can hold the data
	// itself, but length and checksum as well.
	requiredDataSize := MaxLengthBufferLen + length + MaxChecksumBufferLen
	if uint64(len(r.data)) < requiredDataSize {
		// We increase the data slice by a factor of 1.5 to amortise memory allocations over multiple calls. A naive
		// implementation would do a "requiredDataSize * 3 / 2" to get the desired new size. But that approach runs
		// the risk of overflowing the integer when multiplying with 3. What we do instead is, to divide the integer
		// by half by moving all bits right by one bit and adding it to the original integer. That way we achieve a
		// size of 1.5 without overflowing the integer.
		requiredDataSize += requiredDataSize >> 1

		// Round up to the next bigger multiple of 4096 to have buffer sizes aligned with OS page sizes.
		requiredDataSize = (requiredDataSize + 4095) &^ 4095

		newData := make([]byte, requiredDataSize)
		copy(newData, r.data[:lengthBytes])
		r.data = newData
	}
	if _, err := io.ReadFull(r.file, r.data[lengthBytes:uint64(lengthBytes)+length]); err != nil { //nolint:gosec // lengthBytes cannot be negative
		return fmt.Errorf("reading WAL entry data: %w", err)
	}

	// Read the checksum and validate against the data we read so far.
	checksumBytes, err := r.entryChecksumReader(r.file, r.data[uint64(lengthBytes)+length:], r.data[:uint64(lengthBytes)+length]) //nolint:gosec // lengthBytes cannot be negative
	if err != nil {
		return err
	}
	r.value.Data = r.data[lengthBytes : uint64(lengthBytes)+length] //nolint:gosec // lengthBytes cannot be negative
	r.value.SequenceNumber = r.nextSequenceNumber

	r.offset += int64(lengthBytes) + int64(length) + int64(checksumBytes) //nolint:gosec // chances are low that length will overflow
	r.nextSequenceNumber++
	return nil
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

	writerFile, ok := r.file.(SegmentWriterFile)
	if !ok {
		return nil, errors.New("the segment file does not implement the interface for writing to it")
	}

	segmentWriter, err := NewSegmentWriter(writerFile, r.header, r.offset, r.nextSequenceNumber, syncPolicy)
	if err != nil {
		return nil, err
	}

	// Make sure this reader is not used for anything else afterward.
	*r = SegmentReader{}
	return segmentWriter, nil
}

// Close closes the file the SegmentReader is reading from.
func (r *SegmentReader) Close() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	return nil
}
