//go:build windows

package segment

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// renameSegment will rename the segment file by closing it, renaming it and then reopening it again. This is necessary
// on windows, as it does not allow renaming of open files.
func renameSegment(file *os.File, offset int64, newFilePath string) (*os.File, error) {
	oldFilePath := file.Name()
	var err error
	file, err = renameSegmentImpl(file, offset, oldFilePath, newFilePath)
	if err != nil {
		return nil, fmt.Errorf("renaming the WAL segment file from %q to %q: %w", oldFilePath, newFilePath, err)
	}
	return file, nil
}

func renameSegmentImpl(file *os.File, offset int64, oldFilePath string, newFilePath string) (*os.File, error) {
	if err := file.Close(); err != nil {
		return nil, err
	}

	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(newFilePath, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	if _, seekErr := file.Seek(offset, io.SeekStart); seekErr != nil {
		closeErr := file.Close()
		return nil, errors.Join(
			seekErr,
			closeErr,
		)
	}
	return file, nil
}
