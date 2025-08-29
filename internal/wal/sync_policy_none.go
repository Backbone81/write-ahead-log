package wal

// SyncPolicyNone is never flushing the content of the segment to disk. This might improve performance but increases
// the risk of data loss in case of a hardware failure.
type SyncPolicyNone struct {}

// SyncPolicyNone implements SyncPolicy
var _ SyncPolicy = (*SyncPolicyNone)(nil)

func (s *SyncPolicyNone) EntryAppended(sequenceNumber uint64) error {
	return nil
}

func (s *SyncPolicyNone) Close() error {
	return nil
}
