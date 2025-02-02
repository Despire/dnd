//go:build windows

package atomicfile

import (
	"os"
)

// Write writes a file atomically using the temp file technique.
// A temporary file is created, only after succesfully writing to that
// file it will be renamed to the targeted file. The rename operation
// is atomic on POSIX systems.
func Write(name string, contents []byte, perm os.FileMode) (err error) {
	panic("not implemented for windows")
}
