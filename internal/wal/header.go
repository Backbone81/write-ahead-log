package wal

import (
	"encoding/binary"
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

// Write serializes the header and outputs it to the given writer.
func (h *Header) Write(writer io.Writer) error {
	var buffer [HeaderSize]byte
	copy(buffer[:4], h.Magic[:])
	Endian.PutUint16(buffer[4:6], h.Version)
	buffer[6] = byte(h.EntryLengthEncoding)
	buffer[7] = byte(h.EntryChecksumType)
	Endian.PutUint64(buffer[8:16], h.FirstSequenceNumber)
	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("writing WAL header: %w", err)
	}
	return nil
}

// Read deserializes the header from the given reader. It validates the header after reading.
func (h *Header) Read(reader io.Reader) error {
	var buffer [HeaderSize]byte
	if _, err := io.ReadFull(reader, buffer[:]); err != nil {
		return fmt.Errorf("reading WAL header: %w", err)
	}
	if _, err := binary.Decode(buffer[0:4], Endian, h.Magic[:]); err != nil {
		return fmt.Errorf("decoding WAL header Magic bytes: %w", err)
	}
	if _, err := binary.Decode(buffer[4:6], Endian, &h.Version); err != nil {
		return fmt.Errorf("decoding WAL header version: %w", err)
	}
	h.EntryLengthEncoding = EntryLengthEncoding(buffer[6])
	h.EntryChecksumType = EntryChecksumType(buffer[7])
	if _, err := binary.Decode(buffer[8:16], Endian, &h.FirstSequenceNumber); err != nil {
		return fmt.Errorf("decoding WAL header sequence number: %w", err)
	}
	return h.Validate()
}

// Validate makes sure that the header is valid. This includes checking the magic bytes and the version.
func (h *Header) Validate() error {
	if !slices.Equal(h.Magic[:], Magic[:]) {
		return ErrHeaderInvalidMagicBytes
	}
	if h.Version != HeaderVersion {
		return ErrHeaderUnsupportedVersion
	}
	return nil
}
