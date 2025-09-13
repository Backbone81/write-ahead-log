package wal

import (
	"fmt"
	"path"
	"sync"
	"time"
)

// Writer provides the main functionality for writing to the write-ahead log. It abstracts away the fact that the WAL
// is distributed over several segment files and does rollover into new segments as necessary.
//
// Writer is safe to use from multiple Go routines concurrently, as long as the Writer is locked for every access and
// function call. You also need to lock the Writer, even when you work with the WAL in a single Go routine. This is
// necessary, because some sync policies spawn their own Go routines for asynchronous syncs and will take that lock
// to avoid race conditions.
//
// You can only create a writer with the Reader.ToWriter function. This makes sure that you have read all entries before
// writing to the write-ahead log.
type Writer struct {
	sync.Mutex

	segmentWriter *SegmentWriter

	preAllocationSize   int64
	maxSegmentSize      int64
	firstSequenceNumber uint64
	entryLengthEncoding EntryLengthEncoding
	entryChecksumType   EntryChecksumType
	syncPolicy          SyncPolicy
	rolloverCallback    RolloverCallback
}

// RolloverCallback is the callback users can register for getting notified when a rollover of a segment file happens.
// The parameters are the previous and the next segment identified by the first sequence number of entries stored
// inside.
type RolloverCallback func(previousSegment uint64, nextSegment uint64)

// DefaultRolloverCallback provides a callback which does nothing.
var DefaultRolloverCallback RolloverCallback = func(previousSegment uint64, nextSegment uint64) {}

// WriterOption describes the function signature which all writer options need to implement.
type WriterOption func(w *Writer)

// WithPreAllocationSize overwrites the default pre-allocation size of new segment files.
// Can be used with Init and Reader.ToWriter.
func WithPreAllocationSize(preAllocationSize int64) WriterOption {
	return func(w *Writer) {
		w.preAllocationSize = preAllocationSize
	}
}

// WithMaxSegmentSize overwrites the default maximum segment size which causes rollover into a new segment when reached.
// Can be used with Reader.ToWriter.
func WithMaxSegmentSize(maxSegmentSize int64) WriterOption {
	return func(w *Writer) {
		w.maxSegmentSize = maxSegmentSize
	}
}

// WithEntryLengthEncoding overwrites the default entry length encoding.
// Can be used with Init and Reader.ToWriter.
func WithEntryLengthEncoding(entryLengthEncoding EntryLengthEncoding) WriterOption {
	return func(w *Writer) {
		w.entryLengthEncoding = entryLengthEncoding
	}
}

// WithEntryChecksumType overwrites the default entry checksum type.
// Can be used with Init and Reader.ToWriter.
func WithEntryChecksumType(entryChecksumType EntryChecksumType) WriterOption {
	return func(w *Writer) {
		w.entryChecksumType = entryChecksumType
	}
}

// WithSyncPolicyNone overwrites the default sync policy with sync policy none.
// Can be used with Reader.ToWriter.
func WithSyncPolicyNone() WriterOption {
	return func(w *Writer) {
		w.syncPolicy = NewSyncPolicyNone()
	}
}

// WithSyncPolicyImmediate overwrites the default sync policy with sync policy immediate.
// Can be used with Reader.ToWriter.
func WithSyncPolicyImmediate() WriterOption {
	return func(w *Writer) {
		w.syncPolicy = NewSyncPolicyImmediate()
	}
}

// WithSyncPolicyPeriodic overwrites the default sync policy with sync policy periodic.
// Can be used with Reader.ToWriter.
func WithSyncPolicyPeriodic(syncAfterEntryCount int, syncEvery time.Duration) WriterOption {
	return func(w *Writer) {
		w.syncPolicy = NewSyncPolicyPeriodic(syncAfterEntryCount, syncEvery, &w.Mutex)
	}
}

// WithSyncPolicyGrouped overwrites the default sync policy with sync policy grouped.
// Can be used with Reader.ToWriter.
func WithSyncPolicyGrouped(syncAfter time.Duration) WriterOption {
	return func(w *Writer) {
		w.syncPolicy = NewSyncPolicyGrouped(syncAfter, &w.Mutex)
	}
}

// WithRolloverCallback sets the given callback for being triggered when the current segment is rolled.
// Can be used with Reader.ToWriter.
func WithRolloverCallback(rolloverCallback RolloverCallback) WriterOption {
	return func(w *Writer) {
		w.rolloverCallback = rolloverCallback
	}
}

// FilePath returns the file path of the file this writer is writing to.
func (w *Writer) FilePath() string {
	return w.segmentWriter.FilePath()
}

// Header returns the segment file header.
func (w *Writer) Header() Header {
	return w.segmentWriter.Header()
}

// Offset returns the offset in bytes from the start of the file.
func (w *Writer) Offset() int64 {
	return w.segmentWriter.Offset()
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (w *Writer) NextSequenceNumber() uint64 {
	return w.segmentWriter.NextSequenceNumber()
}

// AppendEntry appends the given data as a new entry to the write-ahead log. It will roll over to the next segment
// file before appending if the current file size exceeds the desired maximum segment size.
func (w *Writer) AppendEntry(data []byte) error {
	if err := w.rolloverIfNeeded(); err != nil {
		return err
	}
	if _, err := w.segmentWriter.AppendEntry(data); err != nil {
		return fmt.Errorf("writing entry to segment file: %w", err)
	}
	return nil
}

// Close closes the underlying writer.
func (w *Writer) Close() error {
	return w.segmentWriter.Close()
}

// rolloverIfNeeded will check if the current offset exceeds the desired maximum segment size and do a rollover then.
func (w *Writer) rolloverIfNeeded() error {
	if w.segmentWriter.Offset() < w.maxSegmentSize {
		// We did not yet reach the desired maximum segment size. We can continue with what we have at hand.
		return nil
	}

	return w.rollover()
}

// rollover closes the current writer and creates a new segment to write to.
func (w *Writer) rollover() error {
	previousSegment := w.segmentWriter.Header().FirstSequenceNumber
	if err := w.segmentWriter.Close(); err != nil {
		return err
	}

	nextSegmentWriter, err := CreateSegment(path.Dir(w.segmentWriter.FilePath()), w.segmentWriter.NextSequenceNumber(), CreateSegmentConfig{
		PreAllocationSize:   w.preAllocationSize,
		EntryLengthEncoding: w.entryLengthEncoding,
		EntryChecksumType:   w.entryChecksumType,
		SyncPolicy:          w.syncPolicy,
	})
	if err != nil {
		return err
	}

	w.segmentWriter = nextSegmentWriter
	nextSegment := w.segmentWriter.Header().FirstSequenceNumber
	w.rolloverCallback(previousSegment, nextSegment)
	return nil
}
