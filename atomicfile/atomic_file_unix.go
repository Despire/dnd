//go:build !windows

package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write writes a file atomically using the temp file technique.
// A temporary file is created, only after succesfully writing to that
// file it will be renamed to the targeted file. The rename operation
// is atomic on POSIX systems.
func Write(name string, contents []byte, perm os.FileMode) (err error) {
	directory := filepath.Dir(name)
	tempFile := filepath.Base(name) + "_temp"

	fd, err := os.CreateTemp(directory, tempFile)
	if err != nil {
		return err
	}

	tmp := fd.Name()
	defer func() {
		if err != nil {
			fd.Close()
			os.Remove(tmp)
		}
	}()

	if err = fd.Chmod(perm); err != nil {
		return err
	}

	n, err := fd.Write(contents)
	if err != nil {
		return err
	}

	if n != len(contents) {
		err = fmt.Errorf("failed to fully write contents to tmp file %s: [%v/%v]", tmp, n, len(contents))
		return
	}

	// commit to disk
	if err = fd.Sync(); err != nil {
		return err
	}

	if err = fd.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmp, name); err != nil {
		return err
	}

	return nil
}
