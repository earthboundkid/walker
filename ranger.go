// Package Walker: File Ranger contains convenient iterators for using an fs.WalkFunc.
package walker

import (
	"io/fs"
	"iter"
	"path/filepath"
)

// Ranger provides a convenient way to walk through a directory structure.
type Ranger struct {
	fsys                       fs.FS
	root                       string
	walking                    bool
	skipDir                    bool
	lastErr                    error
	includeFiles, excludeFiles FilterFunc
	includeDirs, excludeDirs   FilterFunc
	erp                        ErrorPolicy
}

// includeAll is a default FilterFunc that includes all files and directories.
func includeAll(Entry) bool {
	return true
}

// excludeNone is a default FilterFunc that includes all files and directories.
func excludeNone(Entry) bool {
	return false
}

// New creates a new *Walker instance with the given root directory.
// Pass a nil fs.FS to use filepath.WalkFunc and walk the OS filesystem.
// The default Walker includes all files and directories.
// There is no default ErrorPolicy.
func New(fsys fs.FS, root string, erp ErrorPolicy) Ranger {
	return Ranger{
		fsys:         fsys,
		root:         root,
		includeFiles: includeAll,
		excludeFiles: excludeNone,
		includeDirs:  includeAll,
		excludeDirs:  excludeNone,
		erp:          erp,
	}
}

// Walk returns an iterator that walks the filepath
// while filtering files and directories and following the ErrorPolicy.
func (w *Ranger) Walk() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		for e := range w.walk {
			if w.HasError() {
				if !w.erp(w.Err(), e) {
					return
				}
				continue
			}
			if w.excludeDirs(e) || !w.includeDirs(e) {
				if e.IsDir() {
					w.SkipDir()
				}
				continue
			}

			if w.excludeFiles(e) || !w.includeFiles(e) {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}

// walk is lower level and doesn't know about the error policy or filters
func (w *Ranger) walk(yield func(Entry) bool) {
	if w.walking {
		panic("already walking")
	}
	if w.erp == nil {
		panic("no error policy set")
	}
	var e Entry
	e.useFilepath = w.fsys == nil
	w.walking = true
	walkDir := func(path string, d fs.DirEntry, err error) error {
		e.Path, e.DirEntry, w.lastErr = path, d, err
		if !yield(e) {
			return fs.SkipAll
		}
		if w.skipDir {
			w.skipDir = false
			if e.Path != w.root {
				return fs.SkipDir
			}
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

// Err returns the last error encountered during walking, if any.
func (r *Ranger) Err() error {
	return r.lastErr
}

// HasError returns true if an error has been encountered during the last walk.
func (r *Ranger) HasError() bool {
	return r.Err() != nil
}

// SkipDir signals to a Ranger that the current directory should be skipped.
func (r *Ranger) SkipDir() {
	r.skipDir = true
}

// ErrorPolicy sets the ErrorPolicy associated with the Walker.
func (w *Ranger) ErrorPolicy(erp ErrorPolicy) {
	w.erp = erp
}

// Include tells the Walker to include matching files when iterating.
func (w *Ranger) Include(f FilterFunc) {
	w.includeFiles = f
}

// Exclude tells the Walker to exclude matching files when iterating.
// Files matched by Exclude take precedence over files matched by Include.
func (w *Ranger) Exclude(f FilterFunc) {
	w.excludeFiles = f
}

// IncludeDir tells the Walker to recursing into matching directories.
func (w *Ranger) IncludeDir(f FilterFunc) {
	w.includeDirs = f
}

// ExcludeDir tells the Walker not to recursing into matching directories.
// Directories matched by ExcludeDir take precedence over directories matched by IncludeDir.
func (w *Ranger) ExcludeDir(f FilterFunc) {
	w.excludeDirs = f
}

// Paths returns an iterator of file paths.
func (w *Ranger) Paths(erp ErrorPolicy) iter.Seq[string] {
	return func(yield func(string) bool) {
		for e := range w.Walk() {
			if !yield(e.Path) {
				return
			}
		}
	}
}

// Entries returns a sequence of paths and directory entries in root.
func (w *Ranger) Entries() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for e := range w.Walk() {
			if !yield(e.Path, e.DirEntry) {
				return
			}
		}
	}
}

// Files returns a sequence of paths and directory entries
// for files in root, ignoring directories.
func (w *Ranger) Files() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for path, de := range w.Entries() {
			if !de.IsDir() && !yield(path, de) {
				return
			}
		}
	}
}

// FilePaths returns a sequence of file paths,
// ignoring directories.
func (w *Ranger) FilePaths() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.Files() {
			if !yield(path) {
				return
			}
		}
	}
}

// DirEntries returns an iterator of fs.DirEntry
// while following the ErrorPolicy.
func (w *Ranger) DirEntries() iter.Seq[fs.DirEntry] {
	return func(yield func(fs.DirEntry) bool) {
		for e := range w.Walk() {
			if !yield(e.DirEntry) {
				return
			}
		}
	}
}
