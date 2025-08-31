package wal

import (
	"fmt"
	"os"
)

// SyncPolicyImmediate is flushing the content of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure, but it has a negative impact on performance.
type SyncPolicyImmediate struct {
	file *os.File
}

// SyncPolicyImmediate implements SyncPolicy.
var _ SyncPolicy = (*SyncPolicyImmediate)(nil)

func (s *SyncPolicyImmediate) EntryAppended(sequenceNumber uint64) error {
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("synching the segment file: %w", err)
	}
	return nil
}

func (s *SyncPolicyImmediate) Close() error {
	return nil
}
