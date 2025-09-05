package wal

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// SyncPolicyPeriodic is flushing segments to disk after having written some number of entries, or after some time
// interval has passed.
// Access to this sync policy needs to be synchronized externally. As it starts a go routine, the external mutex
// is saved in the struct to use in the go routine for synchronization.
type SyncPolicyPeriodic struct {
	syncAfterEntryCount int
	syncEvery           time.Duration
	mutex               *sync.Mutex

	file              SegmentWriterFile
	syncTicker        *time.Ticker
	shutdown          chan struct{}
	shutdownWaitGroup sync.WaitGroup

	unsyncedEntryCount int
}

// SyncPolicyPeriodic implements SyncPolicy.
var _ SyncPolicy = (*SyncPolicyPeriodic)(nil)

// NewSyncPolicyPeriodic creates a new SyncPolicyPeriodic.
func NewSyncPolicyPeriodic(syncAfterEntryCount int, syncEvery time.Duration, mutex *sync.Mutex) *SyncPolicyPeriodic {
	return &SyncPolicyPeriodic{
		syncAfterEntryCount: syncAfterEntryCount,
		syncEvery:           syncEvery,
		mutex:               mutex,
	}
}

func (s *SyncPolicyPeriodic) Startup(file SegmentWriterFile) error {
	s.file = file
	s.syncTicker = time.NewTicker(s.syncEvery)
	s.shutdown = make(chan struct{})
	s.shutdownWaitGroup.Add(1)
	go s.backgroundTask()
	return nil
}

func (s *SyncPolicyPeriodic) EntryAppended(sequenceNumber uint64) error {
	s.unsyncedEntryCount++
	if s.unsyncedEntryCount < s.syncAfterEntryCount {
		return nil
	}

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
}

func (s *SyncPolicyPeriodic) Shutdown() error {
	// Shutdown and wait for the periodic sync to exit.
	s.syncTicker.Stop()
	close(s.shutdown)
	s.shutdownWaitGroup.Wait()

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
}

func (s *SyncPolicyPeriodic) Clone() SyncPolicy {
	return &SyncPolicyPeriodic{
		syncAfterEntryCount: s.syncAfterEntryCount,
		syncEvery:           s.syncEvery,
		mutex:               s.mutex,
	}
}

func (s *SyncPolicyPeriodic) String() string {
	return "periodic"
}

func (s *SyncPolicyPeriodic) backgroundTask() {
	defer s.shutdownWaitGroup.Done()
	for {
		select {
		case <-s.syncTicker.C:
			s.periodicSync()
		case <-s.shutdown:
			return
		}
	}
}

func (s *SyncPolicyPeriodic) periodicSync() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.syncNow(); err != nil {
		log.Printf("ERROR: Periodic sync failed: %s\n", err)
		return
	}
}

func (s *SyncPolicyPeriodic) syncNow() error {
	if s.unsyncedEntryCount == 0 {
		return nil
	}

	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("flushing WAL segment file: %w", err)
	}
	s.unsyncedEntryCount = 0
	return nil
}
