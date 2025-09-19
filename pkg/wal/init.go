package wal

import intwal "github.com/backbone81/write-ahead-log/internal/wal"

// IsInitialized reports if there is already a write-ahead log available in the given directory.
var IsInitialized = intwal.IsInitialized

// Init initializes a new write-ahead log in the given directory.
var Init = intwal.Init

// InitIfRequired initializes the write-ahead log if it is not yet initialized.
var InitIfRequired = intwal.InitIfRequired
