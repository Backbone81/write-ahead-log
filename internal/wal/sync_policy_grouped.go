package wal

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// SyncPolicyGrouped is batching multiple changes of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure significantly, but it has a negative impact on performance.
// Access to this sync policy needs to be synchronized externally. As it starts a go routine, the external mutex
// is saved in the struct to use in the go routine for synchronization.
type SyncPolicyGrouped struct {
	syncAfter time.Duration
	mutex     *sync.Mutex

	file              SegmentWriterFile
	syncTimer         *time.Timer
	shutdown          chan struct{}
	shutdownWaitGroup sync.WaitGroup
	backgroundSync    sync.Cond

	pendingSequenceNumber uint64
	syncedSequenceNumber  uint64
	syncTimerActive       bool
}

// SyncPolicyGrouped implements SyncPolicy.
var _ SyncPolicy = (*SyncPolicyGrouped)(nil)

func NewSyncPolicyGrouped(syncAfter time.Duration, mutex *sync.Mutex) *SyncPolicyGrouped {
	return &SyncPolicyGrouped{
		syncAfter: syncAfter,
		mutex:     mutex,
	}
}

func (s *SyncPolicyGrouped) Startup(file SegmentWriterFile) error {
	s.file = file
	s.syncTimer = time.NewTimer(math.MaxInt64)
	s.shutdown = make(chan struct{})
	s.backgroundSync.L = s.mutex
	s.shutdownWaitGroup.Add(1)
	go s.backgroundTask()
	return nil
}

func (s *SyncPolicyGrouped) EntryAppended(sequenceNumber uint64) error {
	if !s.syncTimerActive {
		s.syncTimer.Reset(s.syncAfter)
		s.syncTimerActive = true
	}

	s.pendingSequenceNumber = max(s.pendingSequenceNumber, sequenceNumber)
	for s.syncedSequenceNumber < sequenceNumber {
		s.backgroundSync.Wait()
	}
	return nil
}

func (s *SyncPolicyGrouped) Shutdown() error {
	// Shutdown and wait for the periodic sync to exit.
	s.syncTimer.Stop()
	close(s.shutdown)
	s.shutdownWaitGroup.Wait()

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
}

func (s *SyncPolicyGrouped) Clone() SyncPolicy {
	return &SyncPolicyGrouped{
		syncAfter: s.syncAfter,
		mutex:     s.mutex,
	}
}

func (s *SyncPolicyGrouped) String() string {
	return "grouped"
}

func (s *SyncPolicyGrouped) backgroundTask() {
	defer s.shutdownWaitGroup.Done()
	for {
		select {
		case <-s.syncTimer.C:
			s.timedSync()
		case <-s.shutdown:
			return
		}
	}
}

func (s *SyncPolicyGrouped) timedSync() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.syncTimerActive = false

	if err := s.syncNow(); err != nil {
		log.Printf("ERROR: Timed sync failed: %s\n", err)
		return
	}
}

func (s *SyncPolicyGrouped) syncNow() error {
	if s.syncedSequenceNumber == s.pendingSequenceNumber {
		return nil
	}

	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("flushing WAL segment file: %w", err)
	}
	s.syncedSequenceNumber = s.pendingSequenceNumber
	s.backgroundSync.Broadcast()
	return nil
}
