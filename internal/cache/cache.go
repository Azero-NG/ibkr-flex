package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const subdir = "ibkr-flex"

type Fetcher func() ([]byte, error)

func filePath(queryID string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("cache: locate user cache dir: %w", err)
	}
	day := time.Now().Format("20060102")
	return filepath.Join(base, subdir, fmt.Sprintf("%s-%s.xml", queryID, day)), nil
}

// Get returns cached XML if today's file exists; otherwise calls fetcher and persists.
// When refresh is true, any existing cache file is bypassed and overwritten.
func Get(queryID string, refresh bool, fetcher Fetcher) ([]byte, error) {
	path, err := filePath(queryID)
	if err != nil {
		return nil, err
	}
	if !refresh {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("cache: read %s: %w", path, err)
		}
	}
	data, err := fetcher()
	if err != nil {
		return nil, err
	}
	if err := writeAtomic(path, data); err != nil {
		return nil, fmt.Errorf("cache: write %s: %w", path, err)
	}
	return data, nil
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ibkr-flex-*.xml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}
