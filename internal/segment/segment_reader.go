package segment

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/backbone81/write-ahead-log/internal/encoding"
	"github.com/backbone81/write-ahead-log/internal/utils"
)

var ErrEntryNone = errors.New("this is no WAL entry")

// SegmentReaderFile is an interface which needs to be implemented by the file to read from.
type SegmentReaderFile interface {
	io.ReadCloser
	io.Seeker
	Name() string
}

// SegmentReader provides functionality for reading from a single segment file.
//
// Instances of SegmentReader are NOT safe to use concurrently. You need to provide external synchronization.
type SegmentReader struct {
	noCopy utils.NoCopy

	// The segment file to read from.
	file SegmentReaderFile

	// The header of the segment file.
	header encoding.Header

	// The current offset from the start of the file in bytes. This is used together with fileSize to calculate the
	// available data until the end of the file, and to reset to a former offset after a failed read of an entry.
	offset int64

	// The sequence number the next entry will receive.
	nextSequenceNumber uint64

	// The reader to decode the length of an entry.
	entryLengthReader encoding.EntryLengthReader

	// The reader to calculate and read the checksum.
	entryChecksumReader encoding.EntryChecksumReader

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
// To avoid resources leaking, the returned SegmentReader needs to be closed by calling Shutdown().
// Returns an error if the file cannot be opened, read from or the header is malformed.
func OpenSegment(directory string, firstSequenceNumber uint64) (*SegmentReader, error) {
	segmentFilePath := path.Join(directory, SegmentFileName(firstSequenceNumber))
	segmentReader, err := openSegment(segmentFilePath, firstSequenceNumber)
	if err != nil {
		return nil, fmt.Errorf("the WAL segment file %q: %w", segmentFilePath, err)
	}
	return segmentReader, nil
}

func openSegment(segmentFilePath string, firstSequenceNumber uint64) (*SegmentReader, error) {
	file, err := os.OpenFile(segmentFilePath, os.O_RDWR, 0) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	var buffer [encoding.HeaderSize]byte
	header, err := encoding.ReadHeader(file, buffer[:])
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if header.FirstSequenceNumber != firstSequenceNumber {
		return nil, fmt.Errorf("expected first sequence number to be %d but got %d", firstSequenceNumber, header.FirstSequenceNumber)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading file size: %w", err)
	}

	currOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading file position: %w", err)
	}

	segmentReader, err := NewSegmentReader(file, NewSegmentReaderConfig{
		Header:             header,
		FileSize:           fileInfo.Size(),
		Offset:             currOffset,
		NextSequenceNumber: firstSequenceNumber,
	})
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return segmentReader, nil
}

// NewSegmentReaderConfig is the configuration required for a call to NewSegmentReader.
type NewSegmentReaderConfig struct {
	// Header is the segment file header.
	Header encoding.Header

	// Offset is the current position in bytes from the start of the file.
	Offset int64

	// NextSequenceNumber is the sequence number the next entry will receive.
	NextSequenceNumber uint64

	// FileSize is the total size in bytes of the segment file.
	FileSize int64
}

// NewSegmentReader creates a SegmentReader from a file which is already open.
func NewSegmentReader(file SegmentReaderFile, newSegmentReaderConfig NewSegmentReaderConfig) (*SegmentReader, error) {
	entryLengthReader, err := encoding.GetEntryLengthReader(newSegmentReaderConfig.Header.EntryLengthEncoding)
	if err != nil {
		return nil, err
	}

	entryChecksumReader, err := encoding.GetEntryChecksumReader(newSegmentReaderConfig.Header.EntryChecksumType)
	if err != nil {
		return nil, err
	}

	return &SegmentReader{
		file:                file,
		header:              newSegmentReaderConfig.Header,
		offset:              newSegmentReaderConfig.Offset,
		nextSequenceNumber:  newSegmentReaderConfig.NextSequenceNumber,
		entryLengthReader:   entryLengthReader,
		entryChecksumReader: entryChecksumReader,
		data:                make([]byte, 4*1024), // Pre-allocate the data slice to reduce the number of allocations.
		fileSize:            newSegmentReaderConfig.FileSize,
	}, nil
}

// FilePath returns the file path of the file this reader is reading from.
func (r *SegmentReader) FilePath() string {
	return r.file.Name()
}

// Header returns the segment file header.
func (r *SegmentReader) Header() encoding.Header {
	return r.header
}

// Offset returns the offset in bytes from the start of the file.
func (r *SegmentReader) Offset() int64 {
	return r.offset
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (r *SegmentReader) NextSequenceNumber() uint64 {
	return r.nextSequenceNumber
}

// Next reports if an entry has been successfully read. When it returns true, Err() returns nil and Value() contains
// valid data. When it returns false, Err() contains the error and Value() contains invalid data.
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

	ReadEntryTotal.Inc()
	ReadEntryBytes.Add(float64(len(r.value.Data)))
	return true
}

func (r *SegmentReader) next() error {
	// Read the length of the entry.
	// We use the data slice as scratch space for converting bytes to integers. We assume that the data slice can always
	// hold at least the maximum length encoding. This is true for a pre-allocated data slice.
	length, lengthBytes, err := r.entryLengthReader(r.file, r.data[:encoding.MaxLengthBufferLen])
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
	requiredDataSize := encoding.MaxLengthBufferLen + length + encoding.MaxChecksumBufferLen
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
// Returns ErrEntryNone when no entry could be read. This indicates either a corrupt entry or the end of the written
// entries in the pre-allocated segment file.
// Returns io.EOF when the end of the segment file was reached and no more data could be read. This error is still
// wrapped in ErrEntryNone but can be checked for separately.
func (r *SegmentReader) Err() error {
	return r.err
}

// ToWriter returns a SegmentWriter to append to the open segment file. You must have read all entries of the segment
// before you call this method. Otherwise, it will fail. After a call to ToWriter(), you cannot use the SegmentReader
// anymore.
func (r *SegmentReader) ToWriter() (*SegmentWriter, error) {
	if !errors.Is(r.err, ErrEntryNone) {
		return nil, errors.New("segment needs to be read until the last entry is reached")
	}

	writerFile, ok := r.file.(SegmentWriterFile)
	if !ok {
		return nil, errors.New("the segment file does not implement the interface for writing to it")
	}

	segmentWriter, err := NewSegmentWriter(writerFile, NewSegmentWriterConfig{
		Header:             r.header,
		Offset:             r.offset,
		NextSequenceNumber: r.nextSequenceNumber,
	})
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
