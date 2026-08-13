package localfile

import (
	"fmt"
	"os"
	"path/filepath"
)

func AbsPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("output path is required")
	}
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		abs, err := filepath.Abs(cleaned)
		if err != nil {
			return "", fmt.Errorf("resolve output path: %w", err)
		}
		cleaned = abs
	}
	return cleaned, nil
}

func AtomicWriteFile(path string, body []byte, perm os.FileMode, overwrite bool) error {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
			return fmt.Errorf("refusing to overwrite non-regular output path %s", path)
		}
		if !overwrite {
			return fmt.Errorf("output file already exists: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary output file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set output file permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary output file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}
	return nil
}
