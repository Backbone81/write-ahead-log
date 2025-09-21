//go:build !windows

package segment

import (
	"fmt"
	"os"
)

// renameSegment will rename the segment file while being open. This works on linux but not on windows.
func renameSegment(file *os.File, offset int64, newFilePath string) (*os.File, error) {
	oldFilePath := file.Name()
	if err := renameSegmentImpl(oldFilePath, newFilePath); err != nil {
		return nil, fmt.Errorf("renaming the WAL segment file from %q to %q: %w", oldFilePath, newFilePath, err)
	}
	return file, nil
}

func renameSegmentImpl(oldFilePath string, newFilePath string) error {
	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		return err
	}
	return nil
}
