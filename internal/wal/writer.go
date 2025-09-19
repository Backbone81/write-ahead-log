package wal

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"write-ahead-log/internal/encoding"
	"write-ahead-log/internal/segment"
)

// Writer provides the main functionality for writing to the write-ahead log. It abstracts away the fact that the WAL
// is distributed over several segment files and does rollover into new segments as necessary.
//
// Writer is safe to use from multiple Go routines concurrently.
//
// You can only create a writer with the Reader.ToWriter function. This makes sure that you have read all entries before
// writing to the write-ahead log.
type Writer struct {
	mutex sync.Mutex

	segmentWriter *segment.SegmentWriter
	syncPolicy    SyncPolicy

	preAllocationSize   int64
	maxSegmentSize      int64
	firstSequenceNumber uint64
	entryLengthEncoding encoding.EntryLengthEncoding
	entryChecksumType   encoding.EntryChecksumType
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
		w.preAllocationSize = max(preAllocationSize, 0)
	}
}

// WithMaxSegmentSize overwrites the default maximum segment size which causes rollover into a new segment when reached.
// Can be used with Reader.ToWriter.
func WithMaxSegmentSize(maxSegmentSize int64) WriterOption {
	return func(w *Writer) {
		// We need to prevent zero entry segments as they would result in duplicate segment file names. We therefore
		// enforce at least one byte more than the header to have at least one entry in each segment.
		w.maxSegmentSize = max(maxSegmentSize, encoding.HeaderSize+1)
	}
}

// WithEntryLengthEncoding overwrites the default entry length encoding.
// Can be used with Init and Reader.ToWriter.
func WithEntryLengthEncoding(entryLengthEncoding encoding.EntryLengthEncoding) WriterOption {
	return func(w *Writer) {
		w.entryLengthEncoding = entryLengthEncoding
	}
}

// WithEntryChecksumType overwrites the default entry checksum type.
// Can be used with Init and Reader.ToWriter.
func WithEntryChecksumType(entryChecksumType encoding.EntryChecksumType) WriterOption {
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
		w.syncPolicy = NewSyncPolicyPeriodic(syncAfterEntryCount, syncEvery)
	}
}

// WithSyncPolicyGrouped overwrites the default sync policy with sync policy grouped.
// Can be used with Reader.ToWriter.
func WithSyncPolicyGrouped(syncAfter time.Duration) WriterOption {
	return func(w *Writer) {
		w.syncPolicy = NewSyncPolicyGrouped(syncAfter)
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
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// We cut the .new suffix, because during rollover the new segment is created with .new at the end and then renamed
	// to the correct name after the header has been written. This results in the os.File.Name() still returning the
	// not existing .new path. Instead of closing, re-opening and seeking the same file again, we opt for cutting the
	// suffix and let it look like the correct name. This is faster than re-opening the file.
	return strings.TrimSuffix(w.segmentWriter.FilePath(), ".new")
}

// Header returns the segment file header.
func (w *Writer) Header() encoding.Header {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.segmentWriter.Header()
}

// Offset returns the offset in bytes from the start of the file.
func (w *Writer) Offset() int64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.segmentWriter.Offset()
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (w *Writer) NextSequenceNumber() uint64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.segmentWriter.NextSequenceNumber()
}

// AppendEntry appends the given data as a new entry to the write-ahead log. It will roll over to the next segment
// file before appending if the current file size exceeds the desired maximum segment size.
func (w *Writer) AppendEntry(data []byte) (uint64, error) {
	sequenceNumber, err := w.appendEntry(data)
	if err != nil {
		return 0, err
	}

	// Note that the call to the sync policy must not happen under the writer lock. The sync policy can block to
	// group several AppendEntry calls. If this call would happen under the writer lock, we would not be able to have
	// any concurrency at all.
	if err := w.syncPolicy.EntryAppended(sequenceNumber); err != nil {
		return 0, err
	}
	return sequenceNumber, nil
}

func (w *Writer) appendEntry(data []byte) (uint64, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if err := w.rolloverIfNeeded(); err != nil {
		return 0, err
	}
	sequenceNumber, err := w.segmentWriter.AppendEntry(data)
	if err != nil {
		return 0, fmt.Errorf("writing entry to segment file: %w", err)
	}
	return sequenceNumber, nil
}

// Close closes the underlying writer.
func (w *Writer) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	syncErr := w.syncPolicy.Shutdown()
	closeErr := w.segmentWriter.Close()

	return errors.Join(syncErr, closeErr)
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
	RolloverTotal.Inc()
	start := time.Now()

	previousSegment := w.segmentWriter.Header().FirstSequenceNumber

	if err := w.syncPolicy.Shutdown(); err != nil {
		return err
	}
	if err := w.segmentWriter.Truncate(); err != nil {
		return err
	}
	if err := w.segmentWriter.Close(); err != nil {
		return err
	}

	nextSegmentWriter, err := segment.CreateSegment(path.Dir(w.segmentWriter.FilePath()), w.segmentWriter.NextSequenceNumber(), segment.CreateSegmentConfig{
		PreAllocationSize:   w.preAllocationSize,
		EntryLengthEncoding: w.entryLengthEncoding,
		EntryChecksumType:   w.entryChecksumType,
	})
	if err != nil {
		return err
	}
	w.segmentWriter = nextSegmentWriter

	if err := w.syncPolicy.Startup(w.segmentWriter); err != nil {
		return err
	}

	nextSegment := w.segmentWriter.Header().FirstSequenceNumber
	w.rolloverCallback(previousSegment, nextSegment)

	duration := time.Since(start).Seconds()
	if duration > 1.0 {
		log.Printf("WARNING: Segment rollover needed %f seconds which is too slow.\n", duration)
	}
	RolloverDuration.Observe(duration)
	return nil
}
