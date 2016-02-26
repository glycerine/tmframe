package tm

import (
	"os"
)

// FileExists returns true if the named path
// exists in the filesystem and is a file (and
// not a directory).
func FileExists(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}

// DirExists returns true if the named path
// is a directly presently in the filesystem.
func DirExists(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return true
	}
	return false
}
