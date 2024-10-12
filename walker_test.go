package walker_test

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/earthboundkid/walker"
)

func TestWalker(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt":            &fstest.MapFile{Data: []byte("file1 content")},
		"file2.log":            &fstest.MapFile{Data: []byte("file2 content")},
		"dir1/file3.txt":       &fstest.MapFile{Data: []byte("file3 content")},
		"dir1/file4.log":       &fstest.MapFile{Data: []byte("file4 content")},
		"dir2/file5.txt":       &fstest.MapFile{Data: []byte("file5 content")},
		"dir2/subdir/file6.go": &fstest.MapFile{Data: []byte("file6 content")},
	}

	tests := []struct {
		name  string
		setup func(*walker.Walker)
		want  string
	}{
		{
			name:  "Walk all files",
			setup: func(w *walker.Walker) {},
			want:  "dir1/file3.txt; dir1/file4.log; dir2/file5.txt; dir2/subdir/file6.go; file1.txt; file2.log",
		},
		{
			name: "Only .txt files",
			setup: func(w *walker.Walker) {
				w.Include(walker.MatchExtension(".txt"))
			},
			want: "dir1/file3.txt; dir2/file5.txt; file1.txt",
		},
		{
			name: "Exclude .log files",
			setup: func(w *walker.Walker) {
				w.Exclude(walker.MatchExtension(".log"))
			},
			want: "dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Only files in dir1",
			setup: func(w *walker.Walker) {
				w.IncludeDirs(walker.MatchGlob("dir1"))
			},
			want: "dir1/file3.txt; dir1/file4.log",
		},
		{
			name: "Exclude dir2",
			setup: func(w *walker.Walker) {
				w.ExcludeDirs(walker.MatchGlob("dir2"))
			},
			want: "dir1/file3.txt; dir1/file4.log; file1.txt; file2.log",
		},
		{
			name: "Only .txt and .go files",
			setup: func(w *walker.Walker) {
				w.Include(walker.MatchRegexpMust(`\.(txt|go)$`))
			},
			want: "dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Not .go files",
			setup: func(w *walker.Walker) {
				w.Include(walker.Not(walker.MatchExtension(".go")))
			},
			want: "dir1/file3.txt; dir1/file4.log; dir2/file5.txt; file1.txt; file2.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := walker.New(testFS, ".")
			tt.setup(&w)

			paths := slices.Collect(w.FilePathsUntilError())
			if got := strings.Join(paths, "; "); got != tt.want {
				t.Errorf("Walker.FilePathsUntilError: want %q, got %q", tt.want, got)
			}

			paths = slices.Collect(w.FilePathsIgnoringErrors())
			if got := strings.Join(paths, "; "); got != tt.want {
				t.Errorf("Walker.FilePathsIgnoringErrors: want %q, got %q", tt.want, got)
			}
		})
	}
}
