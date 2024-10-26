package walker

import (
	"io/fs"
	"path"
	"path/filepath"
)

// Entry is a single iteration, possibly an error, under consideration by the Ranger.
// An Entry is not valid after the current iteration and may be overwritten.
type Entry struct {
	Path        string
	DirEntry    fs.DirEntry
	useFilepath bool
}

// IsDir returns whether the DirEntry is a directory.
func (e *Entry) IsDir() bool {
	if e.DirEntry == nil {
		return false
	}
	return e.DirEntry.IsDir()
}

// Dir returns the directory of the current path being walked.
func (e *Entry) Dir() string {
	if e.IsDir() {
		return e.Path
	}
	if e.useFilepath {
		return filepath.Dir(e.Path)
	}
	return path.Dir(e.Path)
}

// Base returns the last element of Path, typically the filename.
func (e *Entry) Base() string {
	if e.useFilepath {
		return filepath.Base(e.Path)
	}
	return path.Base(e.Path)
}
