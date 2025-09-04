package wal

import "errors"

var ErrSyncPolicyUnsupported = errors.New("unsupported WAL sync policy")

// SyncPolicyType describes the type of sync policy to apply when writing to the segment file.
type SyncPolicyType int

const (
	SyncPolicyTypeNone SyncPolicyType = iota
	SyncPolicyTypeImmediate
	SyncPolicyTypePeriodic
	SyncPolicyTypeGrouped
)

// String returns a string representation of the sync policy type.
func (s SyncPolicyType) String() string {
	switch s {
	case SyncPolicyTypeNone:
		return "none"
	case SyncPolicyTypeImmediate:
		return "immediate"
	case SyncPolicyTypePeriodic:
		return "periodic"
	case SyncPolicyTypeGrouped:
		return "grouped"
	default:
		return "unknown"
	}
}

// SyncPolicyTypes provides a list of supported sync policies. Helpful for writing tests and benchmarks which iterate
// over all possibilities.
var SyncPolicyTypes = []SyncPolicyType{
	SyncPolicyTypeNone,
	SyncPolicyTypeImmediate,
	SyncPolicyTypePeriodic,
	SyncPolicyTypeGrouped,
}

// DefaultSyncPolicy is the sync policy type which should work fine for most use cases.
const DefaultSyncPolicy = SyncPolicyTypeGrouped

// SyncPolicy is the interface every sync policy needs to implement.
type SyncPolicy interface {
	EntryAppended(sequenceNumber uint64) error
	Close() error
}

// GetSyncPolicy returns an instance of the sync policy matching the sync policy type.
func GetSyncPolicy(syncPolicyType SyncPolicyType) (SyncPolicy, error) {
	switch syncPolicyType {
	case SyncPolicyTypeNone:
		return &SyncPolicyNone{}, nil
	case SyncPolicyTypeImmediate:
		return &SyncPolicyImmediate{}, nil
	case SyncPolicyTypePeriodic:
		return &SyncPolicyPeriodic{}, nil
	case SyncPolicyTypeGrouped:
		return &SyncPolicyGrouped{}, nil
	default:
		return nil, ErrSyncPolicyUnsupported
	}
}
