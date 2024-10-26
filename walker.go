// Package Walker: File Ranger contains convenient iterators for using an fs.WalkFunc.
package walker

import (
	"io/fs"
	"iter"
	"path"
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
	erp                        ErrorPolicy
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
// The default Walker includes all files and directories.
// There is no default ErrorPolicy.
func New(fsys fs.FS, root string, erp ErrorPolicy) Walker {
	return Walker{
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
// while following the ErrorPolicy.
func (w *Walker) Walk() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for range w.walk {
			if w.HasError() {
				if !w.erp(w) {
					return
				}
				continue
			}
			if dir := w.Dir(); w.excludeDirs(dir, w.DirEntry()) ||
				!w.includeDirs(dir, w.DirEntry()) {
				if w.IsDir() {
					w.SkipDir()
				}
				continue
			}

			if w.excludeFiles(w.Path(), w.DirEntry()) ||
				!w.includeFiles(w.Path(), w.DirEntry()) {
				continue
			}
			if !yield(w.Path(), w.DirEntry()) {
				return
			}
		}
	}
}

func (w *Walker) walk(yield func() bool) {
	if w.walking {
		panic("already walking")
	}
	if w.erp == nil {
		panic("no error policy set")
	}
	w.walking = true
	walkDir := func(path string, d fs.DirEntry, err error) error {
		w.path, w.d, w.err = path, d, err
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

// Dir returns the directory of the current path being walked.
func (w *Walker) Dir() string {
	if w.IsDir() {
		return w.Path()
	}
	if w.fsys != nil {
		return path.Dir(w.Path())
	}
	return filepath.Dir(w.Path())
}

// DirEntry returns the fs.DirEntry for the current file or directory.
func (w *Walker) DirEntry() fs.DirEntry {
	return w.d
}

// IsDir returns whether the current path is a directory.
func (w *Walker) IsDir() bool {
	if w.d == nil {
		return false
	}
	return w.d.IsDir()
}

// SkipDir signals that the current directory should be skipped.
func (w *Walker) SkipDir() {
	if w.path != w.root {
		w.skipDir = true
	}
}

// HasError returns true if an error has been encountered during the walk.
func (w *Walker) HasError() bool {
	return w.err != nil
}

// ErrorPolicy sets the ErrorPolicy associated with the Walker.
func (w *Walker) ErrorPolicy(erp ErrorPolicy) {
	w.erp = erp
}

// Include tells the Walker to include matching files when iterating.
func (w *Walker) Include(f FilterFunc) {
	w.includeFiles = f
}

// Exclude tells the Walker to exclude matching files when iterating.
// Files matched by Exclude take precedence over files matched by Include.
func (w *Walker) Exclude(f FilterFunc) {
	w.excludeFiles = f
}

// IncludeDir tells the Walker to recursing into matching directories.
func (w *Walker) IncludeDir(f FilterFunc) {
	w.includeDirs = f
}

// ExcludeDir tells the Walker not to recursing into matching directories.
// Directories matched by ExcludeDir take precedence over directories matched by IncludeDir.
func (w *Walker) ExcludeDir(f FilterFunc) {
	w.excludeDirs = f
}

// Paths returns an iterator of file paths.
func (w *Walker) Paths(erp ErrorPolicy) iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.Walk() {
			if !yield(path) {
				return
			}
		}
	}
}

// Files returns a sequence of paths and directory entries
// for files in root, ignoring directories.
func (w *Walker) Files() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for path, de := range w.Walk() {
			if !w.IsDir() && !yield(path, de) {
				return
			}
		}
	}
}

// FilePaths returns a sequence of file paths,
// ignoring directories.
func (w *Walker) FilePaths() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range w.Files() {
			if !yield(path) {
				return
			}
		}
	}
}

// Entries returns an iterator of fs.DirEntry
// while following the ErrorPolicy.
func (w *Walker) Entries() iter.Seq[fs.DirEntry] {
	return func(yield func(fs.DirEntry) bool) {
		for _, entry := range w.Walk() {
			if !yield(entry) {
				return
			}
		}
	}
}
