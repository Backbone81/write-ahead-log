package wal

// SyncPolicy is the interface every sync policy needs to implement.
type SyncPolicy interface {
	// Startup is always called on a sync policy before the file is written to. It can be used for setting up timers or
	// go routines.
	// The file is the file which the sync policy is expected to flush. The policy is expected to store the file
	// internally for later use.
	Startup(file SegmentWriterFile) error

	// EntryAppended is called after every entry has been written to the segment file. The sequence number is the number
	// of the entry which was written. The policy can decide if it wants to flush immediately or start some timer for
	// an asynchronous flush.
	EntryAppended(sequenceNumber uint64) error

	// Shutdown is always called before the segment file is closed for writing. The policy should shut down any go
	// routines it started during Startup.
	Shutdown() error

	// Clone returns a new instance of the policy which has the identical settings but is ready to run Startup again.
	// It is called whenever the current segment file is rolled over and a new segment writer is created.
	Clone() SyncPolicy

	// String returns the name of the sync policy. This is useful for logging or error messages.
	String() string
}
