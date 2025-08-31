package wal

import (
	"errors"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
)

var ErrEntryChecksumTypeUnsupported = errors.New("unsupported entry checksum type")

// MaxChecksumBufferLen is the site of the buffer which is big enough for all supported checksum types.
const MaxChecksumBufferLen = crc64.Size

// EntryChecksumType describes the type of checksum applied to an entry.
type EntryChecksumType int

const (
	EntryChecksumTypeCrc32 EntryChecksumType = iota + 1 // We do not start at 0 to detect missing values.
	EntryChecksumTypeCrc64
)

// DefaultEntryChecksumType is the checksum type which should work fine for most use cases.
const DefaultEntryChecksumType = EntryChecksumTypeCrc32

// EntryChecksumWriter is the function signature which all entry checksum writer callbacks need to implement.
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

// EntryChecksumReader is the function signature which all entry checksum reader callbacks need to implement.
type EntryChecksumReader func(reader io.Reader, buffer []byte, data []byte) error

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
		return fmt.Errorf("writing entry checksum: %w", err)
	}
	return nil
}

func ReadEntryChecksumCrc32(reader io.Reader, buffer []byte, data []byte) error {
	if _, err := io.ReadFull(reader, buffer[:4]); err != nil {
		return fmt.Errorf("reading entry checksum: %w", err)
	}
	checksum := Endian.Uint32(buffer[:4])
	if checksum != crc32.Checksum(data, crc32ChecksumTable) {
		return errors.New("WAL entry checksum mismatch")
	}
	return nil
}

var crc64ChecksumTable = crc64.MakeTable(crc64.ISO)

func WriteEntryChecksumCrc64(writer io.Writer, buffer []byte, data []byte) error {
	Endian.PutUint64(buffer[:8], crc64.Checksum(data, crc64ChecksumTable))
	if _, err := writer.Write(buffer[:8]); err != nil {
		return fmt.Errorf("writing entry checksum: %w", err)
	}
	return nil
}

func ReadEntryChecksumCrc64(reader io.Reader, buffer []byte, data []byte) error {
	if _, err := io.ReadFull(reader, buffer[:8]); err != nil {
		return fmt.Errorf("reading entry checksum: %w", err)
	}
	checksum := Endian.Uint64(buffer[:8])
	if checksum != crc64.Checksum(data, crc64ChecksumTable) {
		return errors.New("WAL entry checksum mismatch")
	}
	return nil
}
