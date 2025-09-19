package wal

import intwal "write-ahead-log/internal/wal"

// Writer provides the main functionality for writing to the write-ahead log. It abstracts away the fact that the WAL
// is distributed over several segment files and does rollover into new segments as necessary.
//
// Writer is safe to use from multiple Go routines concurrently.
//
// You can only create a writer with the Reader.ToWriter function. This makes sure that you have read all entries before
// writing to the write-ahead log.
type Writer = intwal.Writer

// WithPreAllocationSize overwrites the default pre-allocation size of new segment files.
// Can be used with Init and Reader.ToWriter.
var WithPreAllocationSize = intwal.WithPreAllocationSize

// WithMaxSegmentSize overwrites the default maximum segment size which causes rollover into a new segment when reached.
// Can be used with Reader.ToWriter.
var WithMaxSegmentSize = intwal.WithMaxSegmentSize

// WithEntryLengthEncoding overwrites the default entry length encoding.
// Can be used with Init and Reader.ToWriter.
var WithEntryLengthEncoding = intwal.WithEntryLengthEncoding

// WithEntryChecksumType overwrites the default entry checksum type.
// Can be used with Init and Reader.ToWriter.
var WithEntryChecksumType = intwal.WithEntryChecksumType

// WithSyncPolicyNone overwrites the default sync policy with sync policy none.
// Can be used with Reader.ToWriter.
var WithSyncPolicyNone = intwal.WithSyncPolicyNone

// WithSyncPolicyImmediate overwrites the default sync policy with sync policy immediate.
// Can be used with Reader.ToWriter.
var WithSyncPolicyImmediate = intwal.WithSyncPolicyImmediate

// WithSyncPolicyPeriodic overwrites the default sync policy with sync policy periodic.
// Can be used with Reader.ToWriter.
var WithSyncPolicyPeriodic = intwal.WithSyncPolicyPeriodic

// WithSyncPolicyGrouped overwrites the default sync policy with sync policy grouped.
// Can be used with Reader.ToWriter.
var WithSyncPolicyGrouped = intwal.WithSyncPolicyGrouped

// WithRolloverCallback sets the given callback for being triggered when the current segment is rolled.
// Can be used with Reader.ToWriter.
var WithRolloverCallback = intwal.WithRolloverCallback
