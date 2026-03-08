package test

import (
	applog "2_Go/internal/log"
	"os"
	"path/filepath"
	"testing"
)

// TestRotatingWriter verifies that the rotating writer rotates and prunes backup files correctly.
func TestRotatingWriter(t *testing.T) {
	tempFilePath := filepath.Join("log", "test", "test.log")
	maxSizeMB := 1
	maxFiles := 3

	// Ensure the directory for the temporary log file exists
	dir := filepath.Dir(tempFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory for log file: %v", err)
	}

	rw, err := applog.NewRotatingWriter(tempFilePath, maxSizeMB, maxFiles)
	if err != nil {
		t.Fatalf("Failed to create rotatingWriter: %v", err)
	}
	defer func() {
		os.Remove(tempFilePath)
		dir := filepath.Dir(tempFilePath)
		base := filepath.Base(tempFilePath)
		if matches, _ := filepath.Glob(filepath.Join(dir, base+".*")); matches != nil {
			for _, f := range matches {
				os.Remove(f)
			}
		}
		// Remove the directory containing the temporary log file
		os.RemoveAll(filepath.Dir(dir))
	}()

	// Write 1MB+1 byte four times to trigger rotation four times.
	// With maxFiles=3, pruning reduces backups to 3.
	data := make([]byte, 1024*1024+1)
	for i := 0; i < 4; i++ {
		if _, err := rw.Write(data); err != nil {
			t.Fatalf("Failed to write to rotatingWriter: %v", err)
		}
	}

	// Verify rotated backup files are pruned to maxFiles
	matches, err := filepath.Glob(filepath.Join(dir, filepath.Base(tempFilePath)+".*"))
	if err != nil {
		t.Fatalf("Failed to glob rotated files: %v", err)
	}
	if len(matches) != maxFiles {
		t.Errorf("Expected %d rotated files, but got %d", maxFiles, len(matches))
	}

	// The original file should still exist (active log after last rotation)
	if _, err := os.Stat(tempFilePath); os.IsNotExist(err) {
		t.Errorf("Original log file %s should exist after rotation", tempFilePath)
	}
}
