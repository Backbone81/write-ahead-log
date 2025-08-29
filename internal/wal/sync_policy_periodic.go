package wal

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// SyncPolicyPeriodic is flushing segments to disk after having written some number of entries, or after some time
// interval has passed.
type SyncPolicyPeriodic struct {
	file                *os.File
	syncAfterEntryCount int
	syncTicker          *time.Ticker
	shutdown            chan struct{}
	shutdownWaitGroup   sync.WaitGroup

	mutex              sync.Mutex
	unsyncedEntryCount int
}

// SyncPolicyPeriodic implements SyncPolicy
var _ SyncPolicy = (*SyncPolicyPeriodic)(nil)

func NewSyncPolicyPeriodic(file *os.File, syncAfterEntryCount int, syncEvery time.Duration) *SyncPolicyPeriodic {
	syncTicker := time.NewTicker(syncEvery)
	result := SyncPolicyPeriodic{
		file:                file,
		syncAfterEntryCount: syncAfterEntryCount,
		syncTicker:          syncTicker,
		shutdown:            make(chan struct{}),
	}
	result.shutdownWaitGroup.Add(1)
	go result.backgroundTask()
	return &result
}

func (s *SyncPolicyPeriodic) EntryAppended(sequenceNumber uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.unsyncedEntryCount++
	if s.unsyncedEntryCount < s.syncAfterEntryCount {
		return nil
	}

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
}

func (s *SyncPolicyPeriodic) Close() error {
	// Shutdown and wait for the periodic sync to exit.
	s.syncTicker.Stop()
	close(s.shutdown)
	s.shutdownWaitGroup.Wait()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
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
		// TODO: transport error messages to user
		return
	}
}

func (s *SyncPolicyPeriodic) syncNow() error {
	if s.unsyncedEntryCount == 0 {
		return nil
	}

	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("synching the segment file: %w", err)
	}
	s.unsyncedEntryCount = 0
	return nil
}
