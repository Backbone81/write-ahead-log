package wal

import (
	"fmt"
)

// SyncPolicyImmediate is flushing the content of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure, but it has a negative impact on performance.
type SyncPolicyImmediate struct {
	file SegmentWriterFile
}

// SyncPolicyImmediate implements SyncPolicy.
var _ SyncPolicy = (*SyncPolicyImmediate)(nil)

// NewSyncPolicyImmediate returns a new SyncPolicyImmediate.
func NewSyncPolicyImmediate() *SyncPolicyImmediate {
	return &SyncPolicyImmediate{}
}

func (s *SyncPolicyImmediate) Startup(file SegmentWriterFile) error {
	s.file = file
	return nil
}

func (s *SyncPolicyImmediate) EntryAppended(sequenceNumber uint64) error {
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("flushing WAL segment file: %w", err)
	}
	return nil
}

func (s *SyncPolicyImmediate) Shutdown() error {
	return nil
}

func (s *SyncPolicyImmediate) Clone() SyncPolicy {
	return &SyncPolicyImmediate{}
}

func (s *SyncPolicyImmediate) String() string {
	return "immediate"
}
