package wal

import intwal "write-ahead-log/internal/wal"

// Reader provides functionality to read the write-ahead-log. It abstracts away the fact that the write-ahead log is
// split into multiple segments.
//
// Instances of this struct are NOT safe for concurrent use. Either use it on a single Go routine or provide your own
// external synchronization.
type Reader = intwal.Reader

// NewReader creates a new Reader starting at the given sequence number. It will find the segment the sequence number
// belongs to and read all entries up until the requested sequence number.
var NewReader = intwal.NewReader
