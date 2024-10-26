// Package Walker: File Ranger contains convenient iterators for using an fs.WalkFunc.
package walker

import (
	"path/filepath"
	"regexp"
	"strings"
)

// FilterFunc is a function type used to filter files and directories during the walk.
type FilterFunc func(Entry) bool

// MatchRegexp returns true if the path matches the regular expression.
func MatchRegexp(re *regexp.Regexp) FilterFunc {
	return func(e Entry) bool {
		return re.MatchString(e.Path)
	}
}

// MatchRegexpMust compiles re using regexp.MustCompile and passes it to MatchRegexp.
func MatchRegexpMust(re string) FilterFunc {
	return MatchRegexp(regexp.MustCompile(re))
}

// MatchGlob returns true if the path matches any of the glob patterns.
func MatchGlob(patterns ...string) FilterFunc {
	return func(e Entry) bool {
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, e.Path)
			if err == nil && matched {
				return true
			}
		}
		return false
	}
}

// MatchExtension creates a FilterFunc that filters files based on their extensions.
// It returns true if the file has any of the specified extensions.
func MatchExtension(extensions ...string) FilterFunc {
	for i := range extensions {
		extensions[i] = strings.ToLower(extensions[i])
	}
	return func(e Entry) bool {
		ext := strings.ToLower(filepath.Ext(e.Path))
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
	return func(e Entry) bool {
		return strings.HasPrefix(e.Path, prefix)
	}
}

// WithoutPrefix creates a FilterFunc that matches paths not starting with the given prefix.
func WithoutPrefix(prefix string) FilterFunc {
	return func(e Entry) bool {
		return !strings.HasPrefix(e.Path, prefix)
	}
}

// DotFile reports whether a base name begins with a dot.
// As a special case, it allows for a lone "." as the current directory.
func DotFile(e Entry) bool {
	return e.Base() != "." && strings.HasPrefix(e.Base(), ".")
}

// And chains FilterFuncs and returns whether they are all true.
func And(filters ...FilterFunc) FilterFunc {
	return func(e Entry) bool {
		for _, f := range filters {
			if !f(e) {
				return false
			}
		}
		return true
	}
}

// Or chains FilterFuncs and returns whether at least one is true.
func Or(filters ...FilterFunc) FilterFunc {
	return func(e Entry) bool {
		for _, f := range filters {
			if f(e) {
				return true
			}
		}
		return false
	}
}

// Not inverts a FilterFunc.
func Not(f FilterFunc) FilterFunc {
	return func(e Entry) bool {
		return !f(e)
	}
}
