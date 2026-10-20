package wal

import intsegment "github.com/backbone81/write-ahead-log/internal/segment"

// GetSegments returns a list of sequence numbers representing the start of the corresponding segment. The sequence
// numbers are sorted in ascending order.
var GetSegments = intsegment.GetSegments

// SegmentFileName returns the filename for the segment with the given first sequence number.
var SegmentFileName = intsegment.SegmentFileName

// SequenceNumberFromFileName returns the sequence number used for the segment file name.
var SequenceNumberFromFileName = intsegment.SequenceNumberFromFileName

// SegmentFromSequenceNumber returns the segment a sequence number belongs to.
var SegmentFromSequenceNumber = intsegment.SegmentFromSequenceNumber
