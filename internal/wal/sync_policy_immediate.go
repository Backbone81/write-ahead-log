package wal

import (
	"fmt"

	"github.com/backbone81/write-ahead-log/internal/segment"
)

// SyncPolicyImmediate is flushing the content of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure, but it has a negative impact on performance.
type SyncPolicyImmediate struct {
	segmentWriter *segment.SegmentWriter
}

// SyncPolicyImmediate implements SyncPolicy.
var _ SyncPolicy = (*SyncPolicyImmediate)(nil)

// NewSyncPolicyImmediate returns a new SyncPolicyImmediate.
func NewSyncPolicyImmediate() *SyncPolicyImmediate {
	return &SyncPolicyImmediate{}
}

func (s *SyncPolicyImmediate) Startup(segmentWriter *segment.SegmentWriter) error {
	s.segmentWriter = segmentWriter
	return nil
}

func (s *SyncPolicyImmediate) EntryAppended(sequenceNumber uint64) error {
	if err := s.segmentWriter.Sync(); err != nil {
		return fmt.Errorf("flushing WAL segment file: %w", err)
	}
	return nil
}

func (s *SyncPolicyImmediate) Shutdown() error {
	return nil
}

func (s *SyncPolicyImmediate) String() string {
	return "immediate"
}
