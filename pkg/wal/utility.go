package wal

import intsegment "github.com/backbone81/write-ahead-log/internal/segment"

// GetSegments returns a list of sequence numbers representing the start of the corresponding segment. The sequence
// numbers are sorted in ascending order.
var GetSegments = intsegment.GetSegments
