package main

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// rotatingWriter is an io.Writer that appends to a log file and rotates it
// when it exceeds maxBytes. It never truncates existing content.
type rotatingWriter struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	size     int64
	maxBytes int64
	maxFiles int
}

func newRotatingWriter(path string, maxSizeMB, maxFiles int) (*rotatingWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &rotatingWriter{
		file:     f,
		path:     path,
		size:     info.Size(),
		maxBytes: int64(maxSizeMB) * 1024 * 1024,
		maxFiles: maxFiles,
	}, nil
}

func (rw *rotatingWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.size+int64(len(p)) > rw.maxBytes {
		rw.rotate()
	}

	n, err := rw.file.Write(p)
	rw.size += int64(n)
	return n, err
}

func (rw *rotatingWriter) rotate() {
	rw.file.Close()

	backup := rw.path + "." + time.Now().Format("2006-01-02T15-04-05")
	os.Rename(rw.path, backup)

	rw.pruneOldFiles()

	f, _ := os.OpenFile(rw.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	rw.file = f
	rw.size = 0
}

func (rw *rotatingWriter) pruneOldFiles() {
	dir := filepath.Dir(rw.path)
	base := filepath.Base(rw.path)
	matches, _ := filepath.Glob(filepath.Join(dir, base+".*"))
	sort.Strings(matches)
	for len(matches) > rw.maxFiles {
		os.Remove(matches[0])
		matches = matches[1:]
	}
}
