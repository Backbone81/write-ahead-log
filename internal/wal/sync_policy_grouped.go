package wal

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

// SyncPolicyGrouped is batching multiple changes of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure significantly, but it has a negative impact on performance.
type SyncPolicyGrouped struct {
	file              *os.File
	syncAfter         time.Duration
	syncTimer         *time.Timer
	shutdown          chan struct{}
	shutdownWaitGroup sync.WaitGroup
	backgroundSync    sync.Cond

	mutex                 sync.Mutex
	pendingSequenceNumber uint64
	syncedSequenceNumber  uint64
	syncTimerActive       bool
}

func NewSyncPolicyGrouped(file *os.File, syncAfter time.Duration) *SyncPolicyGrouped {
	syncTimer := time.NewTimer(math.MaxInt64)
	newPolicy := SyncPolicyGrouped{
		file:      file,
		syncAfter: syncAfter,
		syncTimer: syncTimer,
		shutdown:  make(chan struct{}),
	}
	newPolicy.backgroundSync.L = &newPolicy.mutex
	newPolicy.shutdownWaitGroup.Add(1)
	go newPolicy.backgroundTask()
	return &newPolicy
}

func (s *SyncPolicyGrouped) EntryAppended(sequenceNumber uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *SyncPolicyGrouped) Close() error {
	// Shutdown and wait for the periodic sync to exit.
	s.syncTimer.Stop()
	close(s.shutdown)
	s.shutdownWaitGroup.Wait()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
}

func (s *SyncPolicyGrouped) backgroundTask() {
	defer s.shutdownWaitGroup.Done()
	for {
		select {
		case <-s.syncTimer.C:
			s.periodicSync()
		case <-s.shutdown:
			return
		}
	}
}

func (s *SyncPolicyGrouped) periodicSync() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.syncNow(); err != nil {
		// TODO: transport error messages to user
		return
	}
	s.syncTimerActive = false
}

func (s *SyncPolicyGrouped) syncNow() error {
	if s.syncedSequenceNumber == s.pendingSequenceNumber {
		return nil
	}

	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("synching the segment file: %w", err)
	}
	s.syncedSequenceNumber = s.pendingSequenceNumber
	s.backgroundSync.Broadcast()
	return nil
}
