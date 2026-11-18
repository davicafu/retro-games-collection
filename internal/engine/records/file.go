package records

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FormatVersion is the version of the files this package reads and writes.
// A file with another version is treated as unreadable.
const FormatVersion = 1

// DefaultDir is where the collection keeps its files: the user's config
// directory, under retro-games.
func DefaultDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "retro-games"), nil
}

// readJSON decodes path into a new T. A missing file gives the zero T. A
// file that does not decode, or has another FormatVersion, is renamed to
// path.bad so a corrupt file never blocks playing, and gives the zero T.
func readJSON[T any](path string, version func(T) int) (T, error) {
	var zero, v T
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return zero, nil
	}
	if err != nil {
		return zero, err
	}
	if json.Unmarshal(data, &v) == nil && version(v) == FormatVersion {
		return v, nil
	}
	if err := os.Rename(path, path+".bad"); err != nil {
		return zero, fmt.Errorf("keep unreadable %s aside: %w", path, err)
	}
	return zero, nil
}

// writeJSON writes v to path through a temporary file renamed into place,
// so a failed write leaves the old file whole.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
