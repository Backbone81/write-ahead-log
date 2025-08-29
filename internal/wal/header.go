package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
)

type Header struct {
	Magic               [4]byte
	Version             uint32
	FirstSequenceNumber uint64
}

var (
	ErrHeaderInvalidMagicBytes  = errors.New("invalid Magic bytes for WAL header")
	ErrHeaderUnsupportedVersion = errors.New("unsupported version for WAL header")
)

var Magic = [4]byte{'W', 'A', 'L', 0}

func (h *Header) Write(writer io.Writer) error {
	var buffer [16]byte
	copy(buffer[:4], h.Magic[:])
	Endian.PutUint32(buffer[4:8], h.Version)
	Endian.PutUint64(buffer[8:16], h.FirstSequenceNumber)
	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("writing WAL header: %w", err)
	}
	return nil
}

func (h *Header) Read(reader io.Reader) error {
	var buffer [16]byte
	if _, err := io.ReadFull(reader, buffer[:]); err != nil {
		return fmt.Errorf("reading WAL header: %w", err)
	}
	if _, err := binary.Decode(buffer[0:4], Endian, h.Magic); err != nil {
		return fmt.Errorf("decoding WAL header Magic bytes: %w", err)
	}
	if _, err := binary.Decode(buffer[4:8], Endian, h.Version); err != nil {
		return fmt.Errorf("decoding WAL header version: %w", err)
	}
	if _, err := binary.Decode(buffer[8:16], Endian, h.FirstSequenceNumber); err != nil {
		return fmt.Errorf("decoding WAL header sequence number: %w", err)
	}
	return nil
}

func (h *Header) Validate() error {
	if !slices.Equal(h.Magic[:], Magic[:]) {
		return ErrHeaderInvalidMagicBytes
	}
	if h.Version != 1 {
		return ErrHeaderUnsupportedVersion
	}
	return nil
}
