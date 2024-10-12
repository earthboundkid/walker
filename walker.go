// Package Walker: File Ranger contains convenient iterators for using an fs.WalkFunc.
package walker

import (
	"io/fs"
	"iter"
	"path/filepath"
)

// Walker provides a convenient way to walk through a directory structure.
type Walker struct {
	fsys                       fs.FS
	root                       string
	walking                    bool
	path                       string
	d                          fs.DirEntry
	err                        error
	skipDir                    bool
	includeFiles, excludeFiles FilterFunc
	includeDirs, excludeDirs   FilterFunc
}

// includeAll is a default FilterFunc that includes all files and directories.
func includeAll(path string, d fs.DirEntry) bool {
	return true
}

// excludeNone is a default FilterFunc that includes all files and directories.
func excludeNone(path string, d fs.DirEntry) bool {
	return false
}

// New creates a new *Walker instance with the given root directory.
// Pass a nil fs.FS to use filepath.WalkFunc and walk the OS filesystem.
func New(fsys fs.FS, root string) Walker {
	return Walker{
		fsys:         fsys,
		root:         root,
		includeFiles: includeAll,
		excludeFiles: excludeNone,
		includeDirs:  includeAll,
		excludeDirs:  excludeNone,
	}
}

// Walk returns a function that walks the filepath.
func (w *Walker) Walk() func(func() bool) {
	return w.walk
}

func (w *Walker) walk(yield func() bool) {
	if w.walking {
		panic("already walking")
	}
	w.walking = true
	walkDir := func(path string, d fs.DirEntry, err error) error {
		w.path, w.d, w.err = path, d, err
		if d.IsDir() {
			if path != "." && (w.excludeDirs(path, d) || !w.includeDirs(path, d)) {
				return filepath.SkipDir
			}
		}
		if w.excludeFiles(path, d) || !w.includeFiles(path, d) {
			return nil
		}
		if !yield() {
			return filepath.SkipAll
		}
		if w.skipDir {
			w.skipDir = false
			return filepath.SkipDir
		}
		return nil
	}
	if w.fsys != nil {
		_ = fs.WalkDir(w.fsys, w.root, walkDir)
	} else {
		_ = filepath.WalkDir(w.root, walkDir)
	}
	w.walking = false
}

// Err returns the current error encountered during walking, if any.
func (w *Walker) Err() error {
	return w.err
}

// Path returns the current file or directory path being walked.
func (w *Walker) Path() string {
	return w.path
}

// DirEntry returns the fs.DirEntry for the current file or directory.
func (w *Walker) DirEntry() fs.DirEntry {
	return w.d
}

// IsDir returns whether the current path is a directory.
func (w *Walker) IsDir() bool {
	return w.d.IsDir()
}

// SkipDir signals that the current directory should be skipped.
func (w *Walker) SkipDir() {
	w.skipDir = true
}

// HasError returns true if an error has been encountered during the walk.
func (w *Walker) HasError() bool {
	return w.err != nil
}

// WalkUntilError returns an iterator that walks until an error is encountered.
func (w *Walker) WalkUntilError() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for range w.Walk() {
			if w.HasError() || !yield(w.Path(), w.DirEntry()) {
				return
			}
		}
	}
}

// WalkIgnoringErrors returns an iterator that walks ignoring any errors encountered.
func (w *Walker) WalkIgnoringErrors() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for range w.Walk() {
			if w.HasError() {
				continue
			}
			if !yield(w.Path(), w.DirEntry()) {
				return
			}
		}
	}
}

// PathsUntilError returns an iterator of file paths, stopping if an error is encountered.
func (w *Walker) PathsUntilError() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.WalkUntilError() {
			if !yield(path) {
				return
			}
		}
	}
}

// PathsIgnoringErrors returns an iterator of paths, ignoring any errors encountered.
func (w *Walker) PathsIgnoringErrors() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.WalkIgnoringErrors() {
			if !yield(path) {
				return
			}
		}
	}
}

// FilePathsUntilError returns an iterator of file paths, ignoring directories and stopping if an error is encountered.
func (w *Walker) FilePathsUntilError() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.PathsUntilError() {
			if w.IsDir() {
				continue
			}
			if !yield(path) {
				return
			}
		}
	}
}

// FilePathsIgnoringErrors returns an iterator of paths, ignoring directories and any errors encountered.
func (w *Walker) FilePathsIgnoringErrors() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.PathsIgnoringErrors() {
			if w.IsDir() {
				continue
			}
			if !yield(path) {
				return
			}
		}
	}
}

// EntriesUntilError returns an iterator of fs.DirEntry, stopping if an error is encountered.
func (w *Walker) EntriesUntilError() iter.Seq[fs.DirEntry] {
	return func(yield func(fs.DirEntry) bool) {
		for _, entry := range w.WalkUntilError() {
			if !yield(entry) {
				return
			}
		}
	}
}

// EntriesIgnoringErrors returns an iterator of fs.DirEntry, ignoring any errors encountered.
func (w *Walker) EntriesIgnoringErrors() iter.Seq[fs.DirEntry] {
	return func(yield func(fs.DirEntry) bool) {
		for _, entry := range w.WalkIgnoringErrors() {
			if !yield(entry) {
				return
			}
		}
	}
}

// Include sets the filter function for including files.
func (w *Walker) Include(f FilterFunc) {
	w.includeFiles = f
}

// Exclude sets the filter function for excluding files.
func (w *Walker) Exclude(f FilterFunc) {
	w.excludeFiles = f
}

// IncludeDirs sets the filter function for including directories.
func (w *Walker) IncludeDirs(f FilterFunc) {
	w.includeDirs = f
}

// ExcludeDirs sets the filter function for excluding directories.
func (w *Walker) ExcludeDirs(f FilterFunc) {
	w.excludeDirs = f
}
