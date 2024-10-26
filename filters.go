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

// MatchGlobPath returns true if the path matches any of the glob patterns.
func MatchGlobPath(patterns ...string) FilterFunc {
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

// MatchGlobName returns true if Entry.Name() matches any of the glob patterns.
func MatchGlobName(patterns ...string) FilterFunc {
	return func(e Entry) bool {
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, e.Name())
			if err == nil && matched {
				return true
			}
		}
		return false
	}
}

// MatchExtension creates a FilterFunc that filters files based on their extensions.
// It returns true if the file has any of the specified extensions.
// It is case insensitive.
func MatchExtension(extensions ...string) FilterFunc {
	for i := range extensions {
		extensions[i] = strings.ToLower(extensions[i])
	}
	return func(e Entry) bool {
		ext := strings.ToLower(e.Ext())
		for _, e := range extensions {
			if e == ext {
				return true
			}
		}
		return false
	}
}

// MatchPrefixPath creates a FilterFunc that matches paths starting with the given prefix.
func MatchPrefixPath(prefix string) FilterFunc {
	return func(e Entry) bool {
		return strings.HasPrefix(e.Path, prefix)
	}
}

// MatchPrefixName creates a FilterFunc
// that matches if Entry.Name() starts with the given prefix.
func MatchPrefixName(prefix string) FilterFunc {
	return func(e Entry) bool {
		return strings.HasPrefix(e.Path, prefix)
	}
}

// MatchDotFile reports whether an Entry.Name() begins with a dot.
var MatchDotFile FilterFunc = MatchPrefixName(".")

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
