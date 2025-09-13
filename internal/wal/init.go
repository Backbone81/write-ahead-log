package wal

// IsInitialized reports if there is already a write-ahead log available in the given directory.
func IsInitialized(directory string) (bool, error) {
	segments, err := GetSegments(directory)
	if err != nil {
		return false, err
	}
	return len(segments) > 0, nil
}

// Init initializes a new write-ahead log in the given directory.
func Init(directory string, options ...WriterOption) error {
	// We use a writer here, to reuse its options. But we do not work with that writer.
	newWriter := Writer{
		preAllocationSize:   DefaultPreAllocationSize,
		maxSegmentSize:      DefaultPreAllocationSize,
		entryLengthEncoding: DefaultEntryLengthEncoding,
		entryChecksumType:   DefaultEntryChecksumType,
		syncPolicy:          NewSyncPolicyImmediate(),
		rolloverCallback:    DefaultRolloverCallback,
	}
	for _, option := range options {
		option(&newWriter)
	}
	segmentWriter, err := CreateSegment(directory, newWriter.firstSequenceNumber, CreateSegmentConfig{
		PreAllocationSize:   newWriter.preAllocationSize,
		EntryLengthEncoding: newWriter.entryLengthEncoding,
		EntryChecksumType:   newWriter.entryChecksumType,
		SyncPolicy:          newWriter.syncPolicy,
	})
	if err != nil {
		return err
	}
	if err := segmentWriter.Close(); err != nil {
		return err
	}
	return nil
}
