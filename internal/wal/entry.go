package wal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

var (
	ErrEntryNone             = errors.New("this is no WAL entry")
)

// WriteEntry writes the given data as a new WAL entry to the writer.
func WriteEntry(writer *bytes.Buffer, data []byte) (int, error) {
	startIndex := writer.Len()

	var buffer [8]byte
	Endian.PutUint64(buffer[:], uint64(len(data)))
	if _, err := writer.Write(buffer[:]); err != nil {
		return 0, fmt.Errorf("writing WAL entry size: %w", err)
	}
	if len(data) > 0 {
		if _, err := writer.Write(data); err != nil {
			return 0, fmt.Errorf("writing WAL entry data: %w", err)
		}
	}
	Endian.PutUint32(buffer[:], crc32.ChecksumIEEE(writer.Bytes()[startIndex:writer.Len()]))
	if _, err := writer.Write(buffer[:4]); err != nil {
		return 0, fmt.Errorf("writing WAL entry checksum: %w", err)
	}
	return writer.Len() - startIndex, nil
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
func ReadEntry(reader io.Reader, data []byte, maxLength int64) (int, []byte, error) {
	n, data, err := readEntry(reader, data, maxLength)
	if err != nil {
		return 0, nil, errors.Join(ErrEntryNone, err)
	}
	return n, data, nil
}

func readEntry(reader io.Reader, data []byte, maxLength int64) (int, []byte, error) {
	// Read the length of the entry first and validate against the maximum possible length.
	var lengthBuffer [8]byte
	if _, err := io.ReadFull(reader, lengthBuffer[:]); err != nil {
		return 0, data, fmt.Errorf("reading WAL entry size: %w", err)
	}
	length := int64(Endian.Uint64(lengthBuffer[:])) //nolint:gosec

	hash := crc32.NewIEEE()
	if _, err := hash.Write(lengthBuffer[:]); err != nil {
		return 0, nil, err
	}

	if maxLength < length {
		return 0, data, errors.New("the WAL entry data exceeds the maximum possible size")
	}

	// Read the data of the entry and use the data slice provided or re-allocate to a fitting size.
	if int64(cap(data)) < length {
		data = make([]byte, 0, length)
	}
	data = data[:length]
	if _, err := io.ReadFull(reader, data); err != nil {
		return 0, data, fmt.Errorf("reading WAL entry data: %w", err)
	}
	if _, err := hash.Write(data); err != nil {
		return 0, nil, err
	}

	// Read the checksum of the entry and validate against the real data.
	var checksum uint32
	if err := binary.Read(reader, Endian, &checksum); err != nil {
		return 0, data, fmt.Errorf("reading WAL entry checksum: %w", err)
	}
	if checksum != hash.Sum32() {
		return 0, data, ErrEntryChecksumMismatch
	}
	return 8 + len(data) + 4, data, nil
}
