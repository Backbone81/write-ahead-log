package wal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"write-ahead-log/internal/utils"
)

// SegmentWriterFile is an interface which needs to be implemented by the file to write to.
type SegmentWriterFile interface {
	io.WriteCloser
	Name() string
	Sync() error
}

// SegmentWriter provides functionality for writing to a single segment file.
//
// Instances of SegmentWriter are NOT safe to use concurrently. You need to provide external synchronization.
type SegmentWriter struct {
	noCopy utils.NoCopy

	// The segment file to write to.
	file SegmentWriterFile

	// The header of the segment file.
	header Header

	// The current offset in bytes from the start of the file.
	offset int64

	// The sequence number the next entry will receive.
	nextSequenceNumber uint64

	// The writer to encode the length of an entry.
	entryLengthWriter EntryLengthWriter

	// The writer to calculate and write the checksum.
	entryChecksumWriter EntryChecksumWriter

	// This is a temporary buffer for converting integers into slices of bytes. This helps us with reducing the amount
	// of memory allocations.
	scratchBuffer [max(MaxLengthBufferLen, MaxChecksumBufferLen)]byte

	// This buffer is used to combine multiple individual file write commands into a single one to improve performance.
	writeBuffer *bytes.Buffer
}

// CreateSegmentConfig is the configuration required for a call to CreateSegment.
type CreateSegmentConfig struct {
	// PreAllocationSize is the number of bytes the new segment should be in size. Pre-allocation helps to avoid
	// fragmentation on disk and reduces the overhead for growing the file on each individual write.
	PreAllocationSize int64

	// EntryLengthEncoding is the encoding of entry lengths.
	EntryLengthEncoding EntryLengthEncoding

	// EntryChecksumType is the type of entry checksum to use.
	EntryChecksumType EntryChecksumType
}

// DefaultPreAllocationSize is a segment size which should work well for most use cases.
const DefaultPreAllocationSize = 64 * 1024 * 1024

// CreateSegment creates a new segment file in the given directory. It will create the new file with the file extension
// ".new" appended to the file name and rename it after the header has been written to. This ensures that the new
// segment file is only visible in the directory when the header was correctly written and flushed to stable storage.
//
// directory is the directory all segment files are located in.
// firstSequenceNumber is used for deriving the file name and for storing it in the segment header.
// createSegmentConfig provides more configuration for the new segment.
func CreateSegment(directory string, firstSequenceNumber uint64, createSegmentConfig CreateSegmentConfig) (*SegmentWriter, error) {
	// Remove any temporary segment file which might be there from an earlier failure.
	newSegmentFileName := SegmentFileName(firstSequenceNumber) + ".new"
	newSegmentFilePath := path.Join(directory, newSegmentFileName)
	if err := os.Remove(newSegmentFilePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("removing the WAL segment file %q: %w", newSegmentFilePath, err)
	}

	// Create the temporary segment file and pre-allocate its size.
	file, err := os.OpenFile(newSegmentFilePath, os.O_RDWR|os.O_CREATE, 0o664) //nolint:gosec // We can not validate paths in a library.
	if err != nil {
		return nil, fmt.Errorf("creating the WAL segment file %q: %w", newSegmentFilePath, err)
	}
	if createSegmentConfig.PreAllocationSize > 0 {
		if err := file.Truncate(createSegmentConfig.PreAllocationSize); err != nil {
			return nil, fmt.Errorf("pre-allocating the WAL segment file %q: %w", newSegmentFilePath, err)
		}
	}

	// Write the header to the segment file and flush the content to stable storage.
	header := Header{
		Magic:               Magic,
		Version:             HeaderVersion,
		EntryLengthEncoding: createSegmentConfig.EntryLengthEncoding,
		EntryChecksumType:   createSegmentConfig.EntryChecksumType,
		FirstSequenceNumber: firstSequenceNumber,
	}
	var buffer [HeaderSize]byte
	if err := WriteHeader(file, buffer[:], header); err != nil {
		return nil, fmt.Errorf("writing WAL header to segment file %q: %w", newSegmentFilePath, err)
	}
	if err := file.Sync(); err != nil {
		return nil, fmt.Errorf("flushing the WAL segment file %q: %w", newSegmentFilePath, err)
	}

	// Rename the temporary segment file to the final one.
	segmentFilePath := path.Join(directory, SegmentFileName(firstSequenceNumber))
	if err := os.Rename(newSegmentFilePath, segmentFilePath); err != nil {
		return nil, fmt.Errorf("renaming the WAL segment file from %q to %q: %w", newSegmentFilePath, segmentFilePath, err)
	}

	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("reading WAL segment file position: %w", err)
	}

	return NewSegmentWriter(file, NewSegmentWriterConfig{
		Header:             header,
		Offset:             offset,
		NextSequenceNumber: firstSequenceNumber,
	})
}

// NewSegmentWriterConfig is the configuration required for a call to NewSegmentWriter.
type NewSegmentWriterConfig struct {
	// Header is the segment file header.
	Header Header

	// Offset is the current position in bytes from the start of the file.
	Offset int64

	// NextSequenceNumber is the sequence number the next entry will receive.
	NextSequenceNumber uint64
}

// NewSegmentWriter creates a SegmentWriter from a file which is already open.
func NewSegmentWriter(file SegmentWriterFile, newSegmentWriterConfig NewSegmentWriterConfig) (*SegmentWriter, error) {
	entryLengthWriter, err := GetEntryLengthWriter(newSegmentWriterConfig.Header.EntryLengthEncoding)
	if err != nil {
		return nil, err
	}

	entryChecksumWriter, err := GetEntryChecksumWriter(newSegmentWriterConfig.Header.EntryChecksumType)
	if err != nil {
		return nil, err
	}

	return &SegmentWriter{
		file:                file,
		header:              newSegmentWriterConfig.Header,
		offset:              newSegmentWriterConfig.Offset,
		nextSequenceNumber:  newSegmentWriterConfig.NextSequenceNumber,
		entryLengthWriter:   entryLengthWriter,
		entryChecksumWriter: entryChecksumWriter,
		writeBuffer:         bytes.NewBuffer(make([]byte, 0, 4*1024)),
	}, nil
}

// FilePath returns the file path of the file this writer is writing to.
func (w *SegmentWriter) FilePath() string {
	return w.file.Name()
}

// Header returns the segment file header.
func (w *SegmentWriter) Header() Header {
	return w.header
}

// Offset returns the offset in bytes from the start of the file.
func (w *SegmentWriter) Offset() int64 {
	return w.offset
}

// NextSequenceNumber returns the sequence number the next entry will receive.
func (w *SegmentWriter) NextSequenceNumber() uint64 {
	return w.nextSequenceNumber
}

// AppendEntry adds the given entry to the segment.
func (w *SegmentWriter) AppendEntry(data []byte) (uint64, error) {
	AppendEntryTotal.Inc()
	AppendEntryBytes.Add(float64(len(data)))

	w.writeBuffer.Reset()
	if err := w.entryLengthWriter(w.writeBuffer, w.scratchBuffer[:], uint64(len(data))); err != nil {
		return 0, err
	}
	if len(data) > 0 {
		if _, err := w.writeBuffer.Write(data); err != nil {
			return 0, err
		}
	}

	if err := w.entryChecksumWriter(w.writeBuffer, w.scratchBuffer[:], w.writeBuffer.Bytes()); err != nil {
		return 0, err
	}

	if _, err := w.file.Write(w.writeBuffer.Bytes()); err != nil {
		return 0, fmt.Errorf("writing WAL entry to segment file: %w", err)
	}
	sequenceNumber := w.nextSequenceNumber
	w.nextSequenceNumber++
	w.offset += int64(w.writeBuffer.Len())

	return sequenceNumber, nil
}

// Sync flushes the content of the segment to stable storage.
func (w *SegmentWriter) Sync() error {
	SyncTotal.Inc()

	start := time.Now()
	if err := w.file.Sync(); err != nil {
		return err
	}
	duration := time.Since(start).Seconds()
	if duration > 1.0 {
		log.Printf("WARNING: Sync to disk needed %f seconds which is too slow.\n", duration)
	}
	SyncDuration.Observe(duration)
	return nil
}

// Close flushes all pending changes to disk and closes the file.
func (w *SegmentWriter) Close() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	return nil
}
