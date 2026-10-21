package maintenance

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// findSiblingFiles returns the files in dir whose name without extension equals stem,
// e.g. stem "8" matches 8.pdf and 8.png but not 80.pdf. Hidden files are skipped.
func findSiblingFiles(dir, stem string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		if strings.TrimSuffix(name, filepath.Ext(name)) == stem {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	return paths, nil
}

// isWithin reports whether path is inside base.
func isWithin(base, path string) bool {
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// moveToQuarantine moves src into quarantineDir as <objId>_<unix-nano>_<name> and
// returns the new path. Across file systems it copies, checks the size, then removes src.
func moveToQuarantine(src, quarantineDir string, objId int64) (string, error) {
	if err := os.MkdirAll(quarantineDir, 0755); err != nil {
		return "", fmt.Errorf("create quarantine directory: %w", err)
	}

	dst := filepath.Join(quarantineDir, fmt.Sprintf("%d_%d_%s", objId, time.Now().UnixNano(), filepath.Base(src)))
	if _, err := os.Lstat(dst); err == nil {
		return "", fmt.Errorf("quarantine file already exists: %s", dst)
	}

	err := os.Rename(src, dst)
	if err == nil {
		return dst, nil
	}
	if !errors.Is(err, syscall.EXDEV) {
		return "", fmt.Errorf("move file: %w", err)
	}

	if err := copyFile(src, dst); err != nil {
		os.Remove(dst)
		return "", fmt.Errorf("copy file: %w", err)
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return "", err
	}
	if srcInfo.Size() != dstInfo.Size() {
		os.Remove(dst)
		return "", fmt.Errorf("copy size mismatch: %d != %d", srcInfo.Size(), dstInfo.Size())
	}
	if err := os.Remove(src); err != nil {
		return "", fmt.Errorf("remove original after copy: %w", err)
	}
	return dst, nil
}

// copyFile copies a source file to a destination file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	// A failed write can surface only on Close, so its error must be checked.
	if err := out.Close(); err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, si.Mode())
}
