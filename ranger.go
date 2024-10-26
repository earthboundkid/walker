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
func includeAll(Entry) bool { return true }

// excludeNone is a default FilterFunc that includes all files and directories.
func excludeNone(Entry) bool { return false }

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
func (tr *Ranger) Walk() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		for e := range tr.walk {
			if tr.HasError() {
				if !tr.erp(tr.Err(), e) {
					return
				}
				continue
			}
			if tr.excludeDirs(e) || !tr.includeDirs(e) {
				if e.IsDir() {
					tr.SkipDir()
				}
				continue
			}

			if tr.excludeFiles(e) || !tr.includeFiles(e) {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}

// walk is lower level and doesn't know about the error policy or filters
func (tr *Ranger) walk(yield func(Entry) bool) {
	if tr.walking {
		panic("already walking")
	}
	if tr.erp == nil {
		panic("no error policy set")
	}
	var e Entry
	e.useFilepath = tr.fsys == nil
	tr.walking = true
	walkDir := func(path string, d fs.DirEntry, err error) error {
		e.Path, e.DirEntry, tr.lastErr = path, d, err
		if !yield(e) {
			return fs.SkipAll
		}
		if tr.skipDir {
			tr.skipDir = false
			if e.Path != tr.root {
				return fs.SkipDir
			}
		}
		return nil
	}
	if tr.fsys != nil {
		_ = fs.WalkDir(tr.fsys, tr.root, walkDir)
	} else {
		_ = filepath.WalkDir(tr.root, walkDir)
	}
	tr.walking = false
}

// Err returns the last error encountered during walking, if any.
func (tr *Ranger) Err() error {
	return tr.lastErr
}

// HasError returns true if an error has been encountered during the last walk.
func (tr *Ranger) HasError() bool {
	return tr.Err() != nil
}

// SkipDir signals to a Ranger that the current directory should be skipped.
func (tr *Ranger) SkipDir() {
	tr.skipDir = true
}

// ErrorPolicy sets the ErrorPolicy associated with the Walker.
func (tr *Ranger) ErrorPolicy(erp ErrorPolicy) {
	tr.erp = erp
}

// Include tells the Walker to include matching files when iterating.
func (tr *Ranger) Include(f FilterFunc) {
	tr.includeFiles = f
}

// Exclude tells the Walker to exclude matching files when iterating.
// Files matched by Exclude take precedence over files matched by Include.
func (tr *Ranger) Exclude(f FilterFunc) {
	tr.excludeFiles = f
}

// IncludeDir tells the Walker to recursing into matching directories.
func (tr *Ranger) IncludeDir(f FilterFunc) {
	tr.includeDirs = f
}

// ExcludeDir tells the Walker not to recursing into matching directories.
// Directories matched by ExcludeDir take precedence over directories matched by IncludeDir.
func (tr *Ranger) ExcludeDir(f FilterFunc) {
	tr.excludeDirs = f
}

// Paths returns an iterator of file paths.
func (tr *Ranger) Paths(erp ErrorPolicy) iter.Seq[string] {
	return func(yield func(string) bool) {
		for e := range tr.Walk() {
			if !yield(e.Path) {
				return
			}
		}
	}
}

// Entries returns a sequence of paths and directory entries in root.
func (tr *Ranger) Entries() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for e := range tr.Walk() {
			if !yield(e.Path, e.DirEntry) {
				return
			}
		}
	}
}

// Files returns a sequence of paths and directory entries
// for files in root, ignoring directories.
func (tr *Ranger) Files() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for path, de := range tr.Entries() {
			if !de.IsDir() && !yield(path, de) {
				return
			}
		}
	}
}

// FilePaths returns a sequence of file paths,
// ignoring directories.
func (tr *Ranger) FilePaths() iter.Seq[string] {
	return func(yield func(string) bool) {
		for path := range tr.Files() {
			if !yield(path) {
				return
			}
		}
	}
}

// DirEntries returns a sequence of fs.DirEntry.
func (tr *Ranger) DirEntries() iter.Seq[fs.DirEntry] {
	return func(yield func(fs.DirEntry) bool) {
		for e := range tr.Walk() {
			if !yield(e.DirEntry) {
				return
			}
		}
	}
}
