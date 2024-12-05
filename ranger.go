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
	isWalking                  bool
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

// New creates a new *Ranger with the given root directory.
// Pass a nil fsys to use filepath.WalkFunc and walk the OS filesystem instead of an fs.FS.
// The default Ranger includes all files and directories.
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

// Entries returns a sequence of Entries for matching files and directories.
func (tr *Ranger) Entries() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		for e := range tr.walk {
			if tr.HasError() {
				if !tr.erp(tr.Err(), e) {
					return
				}
				continue
			}

			switch {
			case e.Dir() == tr.root && (tr.excludeDirs(e) || !tr.includeDirs(e)):
				continue
			case e.IsDir() && (tr.excludeDirs(e) || !tr.includeDirs(e)):
				tr.SkipDir()
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
	if tr.isWalking {
		panic("already walking")
	}
	if tr.erp == nil {
		panic("no error policy set")
	}
	var e Entry
	e.useFilepath = tr.fsys == nil
	tr.isWalking = true
	walkDir := func(path string, d fs.DirEntry, err error) error {
		e.Path, e.DirEntry, tr.lastErr = path, d, err
		if !yield(e) {
			return fs.SkipAll
		}
		if tr.skipDir {
			tr.skipDir = false
			return fs.SkipDir
		}
		return nil
	}
	if tr.fsys != nil {
		_ = fs.WalkDir(tr.fsys, tr.root, walkDir)
	} else {
		_ = filepath.WalkDir(tr.root, walkDir)
	}
	tr.isWalking = false
}

// Err returns the last error encountered during walking, if any.
func (tr *Ranger) Err() error {
	return tr.lastErr
}

// HasError returns true if an error has been encountered during the last walk.
func (tr *Ranger) HasError() bool {
	return tr.Err() != nil
}

// SkipDir signals to a Ranger during iteration that the current directory should be skipped.
// It is an error to call SkipDir when not iterating.
func (tr *Ranger) SkipDir() {
	if !tr.isWalking {
		panic("SkipDir called when not iterating")
	}
	tr.skipDir = true
}

// Include tells the Ranger to include matching files when iterating.
// The default is to include all files.
func (tr *Ranger) Include(f FilterFunc) {
	tr.includeFiles = f
}

// Exclude tells the Ranger to exclude matching files when iterating.
// Files matched by Exclude take precedence over files matched by Include.
func (tr *Ranger) Exclude(f FilterFunc) {
	tr.excludeFiles = f
}

// IncludeDir tells the Ranger to recursing into matching directories.
// The default is to include all directories.
func (tr *Ranger) IncludeDir(f FilterFunc) {
	tr.includeDirs = f
}

// ExcludeDir tells the Ranger not to recursing into matching directories.
// Directories matched by ExcludeDir take precedence over directories matched by IncludeDir.
func (tr *Ranger) ExcludeDir(f FilterFunc) {
	tr.excludeDirs = f
}

// FilesAndDirs returns a sequence of paths and fs.DirEntries
// for matching files and directories.
func (tr *Ranger) FilesAndDirs() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for e := range tr.Entries() {
			if !yield(e.Path, e.DirEntry) {
				return
			}
		}
	}
}

// FileEntries returns a sequence of Entries for matching files, ignoring directories.
func (tr *Ranger) FileEntries() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		for entry := range tr.Entries() {
			if !entry.IsDir() && !yield(entry) {
				return
			}
		}
	}
}

// Files returns a sequence of paths and fs.DirEntries
// for files in root, ignoring directories.
func (tr *Ranger) Files() iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		for path, de := range tr.FilesAndDirs() {
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
