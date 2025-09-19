package wal

import (
	"fmt"
	"log"
	"sync"
	"time"

	"write-ahead-log/internal/segment"
)

// SyncPolicyGrouped is batching multiple changes of the segment to disk after every entry. This reduces the chances of
// data loss because of hardware failure significantly, but it has a negative impact on performance.
// Access to this sync policy needs to be synchronized externally. As it starts a go routine, the external mutex
// is saved in the struct to use in the go routine for synchronization.
type SyncPolicyGrouped struct {
	mutex sync.Mutex

	syncAfter         time.Duration
	segmentWriter     *segment.SegmentWriter
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

func NewSyncPolicyGrouped(syncAfter time.Duration) *SyncPolicyGrouped {
	return &SyncPolicyGrouped{
		syncAfter: max(syncAfter, 100*time.Microsecond),
	}
}

func (s *SyncPolicyGrouped) Startup(segmentWriter *segment.SegmentWriter) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.segmentWriter = segmentWriter

	// Note that we start the sync timer during startup, even though we do not yet have an append pending. This is
	// necessary to avoid a deadlock during rollover, which is caused by missed appends while the sync policy was
	// shutdown but not yet up. To prevent that, we start the timer immediately, and it will be a no-op when nothing
	// was appended during rollover.
	s.syncTimer = time.NewTimer(s.syncAfter)
	s.syncTimerActive = true

	s.shutdown = make(chan struct{})
	s.backgroundSync.L = &s.mutex
	s.shutdownWaitGroup.Add(1)
	go s.backgroundTask()
	return nil
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

func (s *SyncPolicyGrouped) Shutdown() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.syncTimer.Stop()
	close(s.shutdown)

	// We need to unlock the mutex while waiting for the shutdown, otherwise we run the risk of a deadlock.
	s.mutex.Unlock()
	s.shutdownWaitGroup.Wait()
	s.mutex.Lock()

	if err := s.syncNow(); err != nil {
		return err
	}
	return nil
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

	if err := s.segmentWriter.Sync(); err != nil {
		return fmt.Errorf("flushing WAL segment file: %w", err)
	}
	s.syncedSequenceNumber = s.pendingSequenceNumber
	s.backgroundSync.Broadcast()
	return nil
}
