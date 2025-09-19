package wal

import intencoding "github.com/backbone81/write-ahead-log/internal/encoding"

// EntryChecksumType describes the type of checksum applied to an entry.
type EntryChecksumType = intencoding.EntryChecksumType

const (
	EntryChecksumTypeCrc32 = intencoding.EntryChecksumTypeCrc32
	EntryChecksumTypeCrc64 = intencoding.EntryChecksumTypeCrc64
)
