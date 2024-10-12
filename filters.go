// Package Walker: File Ranger contains convenient iterators for using an fs.WalkFunc.
package walker

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

// FilterFunc is a function type used to filter files and directories during the walk.
type FilterFunc func(path string, d fs.DirEntry) bool

// MatchRegexp creates a FilterFunc that uses a precompiled regular expression to filter paths.
// It returns true if the path matches the regular expression.
func MatchRegexp(re *regexp.Regexp) FilterFunc {
	return func(path string, d fs.DirEntry) bool {
		return re.MatchString(path)
	}
}

// MatchGlob creates a FilterFunc that uses one or more glob patterns to filter paths.
// It returns true if the path matches any of the glob patterns.
func MatchGlob(patterns ...string) FilterFunc {
	return func(path string, d fs.DirEntry) bool {
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, path)
			if err == nil && matched {
				return true
			}
		}
		return false
	}
}

// MatchExtension creates a FilterFunc that filters files based on their extensions.
// It returns true if the file has any of the specified extensions.
// It returns false for directories.
func MatchExtension(extensions ...string) FilterFunc {
	for i := range extensions {
		extensions[i] = strings.ToLower(extensions[i])
	}
	return func(path string, d fs.DirEntry) bool {
		if d.IsDir() {
			return false
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range extensions {
			if e == ext {
				return true
			}
		}
		return false
	}
}

// WithPrefix creates a FilterFunc that matches paths starting with the given prefix.
func WithPrefix(prefix string) FilterFunc {
	return func(path string, d fs.DirEntry) bool {
		return strings.HasPrefix(path, prefix)
	}
}

// WithoutPrefix creates a FilterFunc that matches paths not starting with the given prefix.
func WithoutPrefix(prefix string) FilterFunc {
	return func(path string, d fs.DirEntry) bool {
		return !strings.HasPrefix(path, prefix)
	}
}
