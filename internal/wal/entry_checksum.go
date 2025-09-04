package wal

import (
	"errors"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
)

var (
	ErrEntryChecksumTypeUnsupported = errors.New("unsupported WAL entry checksum type")
	ErrEntryChecksumMismatch        = errors.New("WAL entry checksum mismatch")
)

// MaxChecksumBufferLen is the size of the buffer which is big enough for all supported checksum types.
const MaxChecksumBufferLen = crc64.Size

// EntryChecksumType describes the type of checksum applied to an entry.
type EntryChecksumType int

const (
	EntryChecksumTypeCrc32 EntryChecksumType = iota + 1 // We do not start at 0 to detect missing values.
	EntryChecksumTypeCrc64
)

// String returns a string representation of the checksum.
func (e EntryChecksumType) String() string {
	switch e {
	case EntryChecksumTypeCrc32:
		return "crc32"
	case EntryChecksumTypeCrc64:
		return "crc64"
	default:
		return "unknown" //nolint:goconst
	}
}

// EntryChecksumTypes provides a list of supported checksum types. Helpful for writing tests and benchmarks which
// iterate over all possibilities.
var EntryChecksumTypes = []EntryChecksumType{
	EntryChecksumTypeCrc32,
	EntryChecksumTypeCrc64,
}

// DefaultEntryChecksumType is the checksum type which should work fine for most use cases.
const DefaultEntryChecksumType = EntryChecksumTypeCrc32

// EntryChecksumWriter is the function signature which all entry checksum writer functions need to implement.
// writer is the destination to write the checksum to.
// buffer is a temporary scratch space for converting integers to slices of bytes without having to allocate memory.
// data is the data to actually compute the checksum over.
type EntryChecksumWriter func(writer io.Writer, buffer []byte, data []byte) error

// GetEntryChecksumWriter returns the entry checksum writer function matching the entry checksum type.
func GetEntryChecksumWriter(entryChecksumType EntryChecksumType) (EntryChecksumWriter, error) {
	switch entryChecksumType {
	case EntryChecksumTypeCrc32:
		return WriteEntryChecksumCrc32, nil
	case EntryChecksumTypeCrc64:
		return WriteEntryChecksumCrc64, nil
	default:
		return nil, ErrEntryChecksumTypeUnsupported
	}
}

// EntryChecksumReader is the function signature which all entry checksum reader functions need to implement.
// reader is the source to read the checksum from.
// buffer is a temporary scratch space for converting slices of bytes to integers without having to allocate memory.
// data is the data to calculate the checksum over and compare with the checksum read from reader.
// The return values are the number of bytes read and any error which occurred during reading.
type EntryChecksumReader func(reader io.Reader, buffer []byte, data []byte) (int, error)

// GetEntryChecksumReader returns the entry checksum reader function matching the entry checksum type.
func GetEntryChecksumReader(entryChecksumType EntryChecksumType) (EntryChecksumReader, error) {
	switch entryChecksumType {
	case EntryChecksumTypeCrc32:
		return ReadEntryChecksumCrc32, nil
	case EntryChecksumTypeCrc64:
		return ReadEntryChecksumCrc64, nil
	default:
		return nil, ErrEntryChecksumTypeUnsupported
	}
}

var crc32ChecksumTable = crc32.MakeTable(crc32.IEEE)

// WriteEntryChecksumCrc32 writes the checksum to the writer as uint32.
// The buffer is required to avoid allocations and should be big enough to hold the checksum temporarily.
// The data is the data to calculate the checksum over.
func WriteEntryChecksumCrc32(writer io.Writer, buffer []byte, data []byte) error {
	Endian.PutUint32(buffer[:4], crc32.Checksum(data, crc32ChecksumTable))
	if _, err := writer.Write(buffer[:4]); err != nil {
		return checksumWriteError(err)
	}
	return nil
}

// ReadEntryChecksumCrc32 reads the checksum from the reader as uint32.
// The buffer is required to avoid allocations and should be big enough to hold the checksum temporarily.
// The data is the data to calculate the checksum over and compare to the checksum which was read.
// The return value is the number of bytes read from reader.
func ReadEntryChecksumCrc32(reader io.Reader, buffer []byte, data []byte) (int, error) {
	if n, err := io.ReadFull(reader, buffer[:4]); err != nil {
		return n, checksumReadError(err)
	}
	checksum := Endian.Uint32(buffer[:4])
	if checksum != crc32.Checksum(data, crc32ChecksumTable) {
		return 4, ErrEntryChecksumMismatch
	}
	return 4, nil
}

var crc64ChecksumTable = crc64.MakeTable(crc64.ISO)

// WriteEntryChecksumCrc64 writes the checksum to the writer as uint64.
// The buffer is required to avoid allocations and should be big enough to hold the checksum temporarily.
// The data is the data to calculate the checksum over.
func WriteEntryChecksumCrc64(writer io.Writer, buffer []byte, data []byte) error {
	Endian.PutUint64(buffer[:8], crc64.Checksum(data, crc64ChecksumTable))
	if _, err := writer.Write(buffer[:8]); err != nil {
		return checksumWriteError(err)
	}
	return nil
}

// ReadEntryChecksumCrc64 reads the checksum from the reader as uint64.
// The buffer is required to avoid allocations and should be big enough to hold the checksum temporarily.
// The data is the data to calculate the checksum over and compare to the checksum which was read.
// The return value is the number of bytes read from reader.
func ReadEntryChecksumCrc64(reader io.Reader, buffer []byte, data []byte) (int, error) {
	if n, err := io.ReadFull(reader, buffer[:8]); err != nil {
		return n, checksumReadError(err)
	}
	checksum := Endian.Uint64(buffer[:8])
	if checksum != crc64.Checksum(data, crc64ChecksumTable) {
		return 8, ErrEntryChecksumMismatch
	}
	return 8, nil
}

func checksumWriteError(err error) error {
	return fmt.Errorf("writing WAL entry checksum: %w", err)
}

func checksumReadError(err error) error {
	return fmt.Errorf("reading WAL entry checksum: %w", err)
}
