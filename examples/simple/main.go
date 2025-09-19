package main

import (
	"log"

	"github.com/backbone81/write-ahead-log/pkg/wal"
)

func main() {
	if err := runDemo(); err != nil {
		log.Fatalln(err)
	}
}

func runDemo() error {
	walDir := "."

	// Make sure the write-ahead log is initialized. This means that
	// at least one wal file is in the wal directory.
	if err := wal.InitIfRequired(walDir); err != nil {
		return err
	}

	// This is the sequence number which you store somewhere else. For
	// this example, we always assume 0 but in reality this should
	// change whenever you flush your own state to disk.
	var sequenceNumber uint64 = 0

	// Before we can write to the write-ahead log, we need to read all
	// entries starting from our well-known sequence number.
	log.Println("Opening write-ahead log for reading")
	reader, err := wal.NewReader(walDir, sequenceNumber)
	if err != nil {
		return err
	}
	for reader.Next() {
		// We would use reader.Value().Data and apply it to our saved
		// state. reader.Next will return false when we have reached
		// the end of the write-ahead log.
	}

	// We convert the reader to a writer. Note that this is only
	// possible after having read all entries. Afterward the reader
	// cannot be used anymore.
	log.Println("Opening write-ahead log for writing")
	writer, err := reader.ToWriter()
	if err != nil {
		return err
	}

	// We can now append entries. This can also happen concurrently
	// from multiple go routines.
	log.Println("Appending entry")
	if _, err := writer.AppendEntry([]byte("foo")); err != nil {
		return err
	}

	// At the end we should close the writer to make sure that all
	// data was flushed to stable storage.
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
