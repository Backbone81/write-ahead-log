package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// segmentFileNamePattern is the file pattern all segment files need to follow.
var segmentFileNamePattern = regexp.MustCompile(`\d{20}\.wal`)

// GetSegments returns a list of sequence numbers representing the start of the corresponding segment. The sequence
// numbers are sorted in ascending order.
func GetSegments(directory string) ([]uint64, error) {
	dirEntries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", directory, err)
	}

	result := make([]uint64, 0, 1024)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			// We are not interested in directories.
			continue
		}
		if !segmentFileNamePattern.MatchString(dirEntry.Name()) {
			// We are not interested in files not matching our naming pattern.
			continue
		}
		sequenceNumber, err := strconv.ParseUint(strings.TrimSuffix(dirEntry.Name(), ".wal"), 10, 64)
		if err != nil {
			// This error should never occur when our file name pattern is correct.
			return nil, fmt.Errorf("parsing the sequence number from the file name: %w", err)
		}
		result = append(result, sequenceNumber)
	}

	// The file names returned by os.ReadDir() should already be in the correct order. For additional safety we sort
	// the sequence numbers again, in case the order does not match. Sorting an already sorted list should be a cheap
	// operation.
	slices.Sort(result)
	return result, nil
}

func SegmentFromSequenceNumber(directory string, sequenceNumber uint64) (uint64, error) {
	segments, err := GetSegments(directory)
	if err != nil {
		return 0, err
	}

	index, exact := slices.BinarySearch(segments, sequenceNumber)
	if !exact {
		// If we did not find an exact match, the index is where it would be. So we move back by one to get the segment
		// which would contain the sequence number.
		index -= 1
	}
	if index < 0 {
		// The sequence number is in a segment which does not exist in the directory.
		return 0, fmt.Errorf("no segment available for sequence number %q", sequenceNumber)
	}
	return segments[index], nil
}

func segmentFileName(sequenceNumber uint64) string {
	return fmt.Sprintf("%020d.wal", sequenceNumber)
}

// Endian is the endianness the write-ahead log uses for serializing/deserializing integers to file.
var Endian = binary.LittleEndian
