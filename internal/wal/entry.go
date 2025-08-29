package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

// Entry is a single entry in a segment of the write ahead log.
type Entry struct {
	// Length provides the length of data stored in the entry. It is expected to be identical to len(Data).
	Length int32

	// Data is the data stored in the entry. This can be arbitrary data and is application specific.
	Data []byte

	Checksum uint32
}

var (
	ErrEntryDataSizeMismatch = errors.New("the WAL entry size does not match the data length")
	ErrEntryChecksumMismatch = errors.New("the WAL entry checksum does not match the data")
	ErrEntryNone             = errors.New("this is no WAL entry")
)

// WriteEntry writes the given data as a new WAL entry to the writer.
func WriteEntry(writer io.Writer, data []byte) error {
	var buffer [8]byte
	Endian.PutUint64(buffer[:], uint64(len(data)))
	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("writing WAL entry size: %w", err)
	}
	if len(data) > 0 {
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("writing WAL entry data: %w", err)
		}
	}
	Endian.PutUint32(buffer[:], crc32.ChecksumIEEE(data))
	if _, err := writer.Write(buffer[:4]); err != nil {
		return fmt.Errorf("writing WAL entry checksum: %w", err)
	}
	return nil
}

// ReadEntry reads a single WAL entry from the reader.
//
// The data slice allows you to reduce memory allocations. Give a slice with enough capacity and this function will use
// that slice for reading the data into. If the capacity is not enough, a new slice with enough capacity is allocated.
// If no data slice is provided, a matching slice will always be allocated.
//
// Give maxLength to detect malformed entries early and prevent excessive memory allocations in such situations. Set
// maxLength to the remaining bytes in the current segment.
//
// The byte slice return contains the data of the WAL entry. It will always be the value passed in through the data
// parameter, or a newly allocated slice - even in error situations.
func ReadEntry(reader io.Reader, data []byte, maxLength int64) ([]byte, error) {
	data, err := readEntry(reader, data, maxLength)
	if err != nil {
		return nil, errors.Join(ErrEntryNone, err)
	}
	return data, nil
}

func readEntry(reader io.Reader, data []byte, maxLength int64) ([]byte, error) {
	// Read the length of the entry first and validate against the maximum possible length.
	var length int64
	if err := binary.Read(reader, Endian, &length); err != nil {
		return data, fmt.Errorf("reading WAL entry size: %w", err)
	}
	if maxLength < length {
		return data, fmt.Errorf("the WAL entry data exceeds the maximum possible size")
	}

	// Read the data of the entry and use the data slice provided or re-allocate to a fitting size.
	if int64(cap(data)) < length {
		data = make([]byte, 0, length)
	}
	data = data[:length]
	if _, err := io.ReadFull(reader, data); err != nil {
		return data, fmt.Errorf("reading WAL entry data: %w", err)
	}

	// Read the checksum of the entry and validate against the real data.
	var checksum uint32
	if err := binary.Read(reader, Endian, &checksum); err != nil {
		return data, fmt.Errorf("reading WAL entry checksum: %w", err)
	}
	if checksum != crc32.ChecksumIEEE(data) {
		return data, ErrEntryChecksumMismatch
	}
	return data, nil
}
