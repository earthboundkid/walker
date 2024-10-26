package walker

import (
	"io/fs"
	"path"
	"path/filepath"
)

// Entry is a single path/fs.DirEntry pair yielded by a Ranger.
// It knows whether to use package filepath or package path for its methods.
type Entry struct {
	Path        string
	DirEntry    fs.DirEntry
	useFilepath bool
}

// IsDir returns whether the DirEntry is a directory.
// If DirEntry is nil, it returns false.
func (e Entry) IsDir() bool {
	if e.DirEntry == nil {
		return false
	}
	return e.DirEntry.IsDir()
}

// Dir returns the directory of the Entry.
// Unlike [path.Dir] or [filepath.Dir],
// it knows whether e represents a directory,
// in which case it returns its own path, not its parent directory.
func (e Entry) Dir() string {
	if e.IsDir() {
		return e.Path
	}
	if e.useFilepath {
		return filepath.Dir(e.Path)
	}
	return path.Dir(e.Path)
}

// Base returns the last element of Path, typically the filename.
// See [path.Base] and [filepath.Base].
func (e Entry) Base() string {
	if e.useFilepath {
		return filepath.Base(e.Path)
	}
	return path.Base(e.Path)
}

// Ext returns the file name extension of Path.
// See [path.Ext] and [filepath.Ext].
func (e Entry) Ext() string {
	if e.useFilepath {
		return filepath.Ext(e.Path)
	}
	return path.Ext(e.Path)

}

// Split splits path immediately following the final separator,
// separating it into a directory and file name component.
// See [path.Split] and [filepath.Split].
func (e Entry) Split() (string, string) {
	if e.useFilepath {
		return filepath.Split(e.Path)
	}
	return path.Split(e.Path)
}
