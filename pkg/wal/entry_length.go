package wal

import (
	intencoding "write-ahead-log/internal/encoding"
)

// EntryLengthEncoding describes the way the length of an entry is encoded.
type EntryLengthEncoding = intencoding.EntryLengthEncoding

const (
	EntryLengthEncodingUint16  = intencoding.EntryLengthEncodingUint16
	EntryLengthEncodingUint32  = intencoding.EntryLengthEncodingUint32
	EntryLengthEncodingUint64  = intencoding.EntryLengthEncodingUint64
	EntryLengthEncodingUvarint = intencoding.EntryLengthEncodingUvarint
)
