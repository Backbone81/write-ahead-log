package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"write-ahead-log/internal/utils"
)

var (
	ErrEntryLengthEncodingUnsupported = errors.New("unsupported entry length encoding")
	ErrEntryLengthOverflow            = errors.New("entry length overflow")
)

// MaxLengthBufferLen is the size of the buffer which is big enough for all supported encodings.
const MaxLengthBufferLen = binary.MaxVarintLen64

// EntryLengthEncoding describes the way the length of an entry is encoded.
type EntryLengthEncoding int

const (
	EntryLengthEncodingUint16 EntryLengthEncoding = iota + 1 // We do not start at 0 to detect missing values.
	EntryLengthEncodingUint32
	EntryLengthEncodingUint64
	EntryLengthEncodingUvarint
)

// DefaultEntryLengthEncoding is the length encoding which should work fine for most use cases.
const DefaultEntryLengthEncoding = EntryLengthEncodingUint32

// EntryLengthWriter is the function signature which all entry length writer callbacks need to implement.
type EntryLengthWriter func(writer io.Writer, buffer []byte, length uint64) error

// GetEntryLengthWriter returns the entry length writer function matching the entry length encoding.
func GetEntryLengthWriter(entryLengthEncoding EntryLengthEncoding) (EntryLengthWriter, error) {
	switch entryLengthEncoding {
	case EntryLengthEncodingUint16:
		return WriteEntryLengthUint16, nil
	case EntryLengthEncodingUint32:
		return WriteEntryLengthUint32, nil
	case EntryLengthEncodingUint64:
		return WriteEntryLengthUint64, nil
	case EntryLengthEncodingUvarint:
		return WriteEntryLengthUvarint, nil
	default:
		return nil, ErrEntryLengthEncodingUnsupported
	}
}

// EntryLengthReader is the function signature which all entry length reader callbacks need to implement.
type EntryLengthReader func(reader io.Reader, buffer []byte) (uint64, error)

// GetEntryLengthReader returns the entry length reader function matching the entry length encoding.
func GetEntryLengthReader(entryLengthEncoding EntryLengthEncoding) (EntryLengthReader, error) {
	switch entryLengthEncoding {
	case EntryLengthEncodingUint16:
		return ReadEntryLengthUint16, nil
	case EntryLengthEncodingUint32:
		return ReadEntryLengthUint32, nil
	case EntryLengthEncodingUint64:
		return ReadEntryLengthUint64, nil
	case EntryLengthEncodingUvarint:
		return ReadEntryLengthUvarint, nil
	default:
		return nil, ErrEntryLengthEncodingUnsupported
	}
}

// WriteEntryLengthUint16 writes the length to the writer encoded as uint16.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
// An error is returned when the given length exceeds the maximum possible length.
func WriteEntryLengthUint16(writer io.Writer, buffer []byte, length uint64) error {
	if math.MaxUint16 < length {
		return ErrEntryLengthOverflow
	}

	Endian.PutUint16(buffer[:2], uint16(length))
	if _, err := writer.Write(buffer[:2]); err != nil {
		return fmt.Errorf("writing WAL entry length: %w", err)
	}
	return nil
}

// ReadEntryLengthUint16 reads the length from the reader encoded as uint16.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func ReadEntryLengthUint16(reader io.Reader, buffer []byte) (uint64, error) {
	if _, err := io.ReadFull(reader, buffer[:2]); err != nil {
		return 0, fmt.Errorf("reading WAL entry length: %w", err)
	}
	return uint64(Endian.Uint16(buffer[:2])), nil
}

// WriteEntryLengthUint32 writes the length to the writer encoded as uint32.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
// An error is returned when the given length exceeds the maximum possible length.
func WriteEntryLengthUint32(writer io.Writer, buffer []byte, length uint64) error {
	if math.MaxUint32 < length {
		return ErrEntryLengthOverflow
	}

	Endian.PutUint32(buffer[:4], uint32(length))
	if _, err := writer.Write(buffer[:4]); err != nil {
		return fmt.Errorf("writing entry length: %w", err)
	}
	return nil
}

// ReadEntryLengthUint32 reads the length from the reader encoded as uint32.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func ReadEntryLengthUint32(reader io.Reader, buffer []byte) (uint64, error) {
	if _, err := io.ReadFull(reader, buffer[:4]); err != nil {
		return 0, fmt.Errorf("reading WAL entry length: %w", err)
	}
	return uint64(Endian.Uint32(buffer[:4])), nil
}

// WriteEntryLengthUint64 writes the length to the writer encoded as uint64.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func WriteEntryLengthUint64(writer io.Writer, buffer []byte, length uint64) error {
	Endian.PutUint64(buffer[:8], length)
	if _, err := writer.Write(buffer[:8]); err != nil {
		return fmt.Errorf("writing entry length: %w", err)
	}
	return nil
}

// ReadEntryLengthUint64 reads the length from the reader encoded as uint64.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func ReadEntryLengthUint64(reader io.Reader, buffer []byte) (uint64, error) {
	if _, err := io.ReadFull(reader, buffer[:8]); err != nil {
		return 0, fmt.Errorf("reading WAL entry length: %w", err)
	}
	return Endian.Uint64(buffer[:8]), nil
}

// WriteEntryLengthUvarint writes the length to the writer encoded as uvarint.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func WriteEntryLengthUvarint(writer io.Writer, buffer []byte, length uint64) error {
	n := binary.PutUvarint(buffer[:binary.MaxVarintLen64], length)
	if _, err := writer.Write(buffer[:n]); err != nil {
		return fmt.Errorf("writing entry length: %w", err)
	}
	return nil
}

// ReadEntryLengthUvarint reads the length from the reader encoded as uvarint.
// The buffer is required to avoid allocations and should be big enough to hold the encoded length temporarily.
func ReadEntryLengthUvarint(reader io.Reader, buffer []byte) (uint64, error) {
	myByteReader := utils.NewByteReader(reader, buffer)
	return binary.ReadUvarint(&myByteReader)
}
