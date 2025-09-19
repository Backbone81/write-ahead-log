package wal

import (
	"write-ahead-log/internal/encoding"
	"write-ahead-log/internal/segment"
)

// IsInitialized reports if there is already a write-ahead log available in the given directory.
func IsInitialized(directory string) (bool, error) {
	segments, err := segment.GetSegments(directory)
	if err != nil {
		return false, err
	}
	return len(segments) > 0, nil
}

// Init initializes a new write-ahead log in the given directory.
func Init(directory string, options ...WriterOption) error {
	// We use a writer here, to reuse its options. But we do not work with that writer.
	newWriter := Writer{
		preAllocationSize:   segment.DefaultPreAllocationSize,
		maxSegmentSize:      segment.DefaultPreAllocationSize,
		entryLengthEncoding: encoding.DefaultEntryLengthEncoding,
		entryChecksumType:   encoding.DefaultEntryChecksumType,
		syncPolicy:          NewSyncPolicyImmediate(),
		rolloverCallback:    DefaultRolloverCallback,
	}
	for _, option := range options {
		option(&newWriter)
	}
	segmentWriter, err := segment.CreateSegment(directory, newWriter.firstSequenceNumber, segment.CreateSegmentConfig{
		PreAllocationSize:   newWriter.preAllocationSize,
		EntryLengthEncoding: newWriter.entryLengthEncoding,
		EntryChecksumType:   newWriter.entryChecksumType,
	})
	if err != nil {
		return err
	}
	if err := segmentWriter.Close(); err != nil {
		return err
	}
	return nil
}

// InitIfRequired initializes the write-ahead log if it is not yet initialized.
func InitIfRequired(directory string) error {
	initialized, err := IsInitialized(directory)
	if err != nil {
		return err
	}

	if initialized {
		return nil
	}

	if err := Init(directory); err != nil {
		return err
	}
	return nil
}
