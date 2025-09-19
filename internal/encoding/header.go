package encoding

import (
	"errors"
	"fmt"
	"io"
	"slices"
)

var (
	ErrHeaderInvalidMagicBytes  = errors.New("invalid WAL header magic bytes")
	ErrHeaderUnsupportedVersion = errors.New("unsupported WAL header version")
)

// Header describes the segment file header which is located at the start of every segment file.
type Header struct {
	// These are the magic bytes to identify a segment file. This must always be "WAL" followed by a zero value byte.
	// Encoded as four bytes.
	Magic [4]byte

	// The version of the segment file format to use. This allows us to evolve the header and the file format over
	// time if necessary. Encoded as two bytes.
	Version uint16

	// Describes the way the entry length is encoded in the segment file. Encoded as a single byte.
	EntryLengthEncoding EntryLengthEncoding

	// Describes the way the entry checksum is encoded in the segment file. Encoded as a single byte.
	EntryChecksumType EntryChecksumType

	// The sequence number the first entry in the segment file has. Note that the file name and this header value
	// should always match. To have the first sequence number stored in the header makes it possible to detect
	// accidental file renames.
	// Encoded as eight bytes.
	FirstSequenceNumber uint64
}

// HeaderSize provides the size in bytes of the header. Helpful for reading the full header before decoding individual
// elements.
const HeaderSize = 4 + 2 + 1 + 1 + 8

// Magic holds the magic bytes expected at the start of the file.
var Magic = [4]byte{'W', 'A', 'L', 0}

// HeaderVersion provides the currently supported header version.
const HeaderVersion = 1

// DefaultHeader provides a header configuration which is a sane default in most situations.
var DefaultHeader = Header{
	Magic:               Magic,
	Version:             HeaderVersion,
	EntryLengthEncoding: DefaultEntryLengthEncoding,
	EntryChecksumType:   DefaultEntryChecksumType,
	FirstSequenceNumber: 0,
}

// WriteHeader writes the segment header to the writer.
// The buffer is required to avoid allocations and should be big enough to hold the full header temporarily.
func WriteHeader(writer io.Writer, buffer []byte, header Header) error {
	copy(buffer[:4], header.Magic[:])
	Endian.PutUint16(buffer[4:6], header.Version)
	buffer[6] = byte(header.EntryLengthEncoding)
	buffer[7] = byte(header.EntryChecksumType)
	Endian.PutUint64(buffer[8:16], header.FirstSequenceNumber)
	if _, err := writer.Write(buffer); err != nil {
		return headerWriteError(err)
	}
	return nil
}

// ReadHeader reads the segment header from the reader.
// The buffer is required to avoid allocations and should be big enough to hold the full header temporarily.
// An error is returned when the header does not match expectations (like magic bytes, version, etc.).
func ReadHeader(reader io.Reader, buffer []byte) (Header, error) {
	var result Header
	if _, err := io.ReadFull(reader, buffer[:HeaderSize]); err != nil {
		return Header{}, headerReadError(err)
	}

	copy(result.Magic[:], buffer[:4])
	result.Version = Endian.Uint16(buffer[4:6])
	result.EntryLengthEncoding = EntryLengthEncoding(buffer[6])
	result.EntryChecksumType = EntryChecksumType(buffer[7])
	result.FirstSequenceNumber = Endian.Uint64(buffer[8:16])

	if result.Magic != Magic {
		return Header{}, ErrHeaderInvalidMagicBytes
	}
	if result.Version != HeaderVersion {
		return Header{}, ErrHeaderUnsupportedVersion
	}
	if !slices.Contains(EntryLengthEncodings, result.EntryLengthEncoding) {
		return Header{}, ErrEntryLengthEncodingUnsupported
	}
	if !slices.Contains(EntryChecksumTypes, result.EntryChecksumType) {
		return Header{}, ErrEntryChecksumTypeUnsupported
	}
	return result, nil
}

func headerWriteError(err error) error {
	return fmt.Errorf("writing WAL header: %w", err)
}

func headerReadError(err error) error {
	return fmt.Errorf("reading WAL header: %w", err)
}
