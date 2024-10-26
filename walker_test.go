package walker_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/carlmjohnson/be"
	"github.com/earthboundkid/walker"
)

func TestWalker(t *testing.T) {
	testFS := fstest.MapFS{
		"a.txt":                &fstest.MapFile{},
		"dir1/file3.txt":       &fstest.MapFile{},
		"dir1/file4.log":       &fstest.MapFile{},
		"dir2/file5.txt":       &fstest.MapFile{},
		"dir2/subdir/file6.go": &fstest.MapFile{},
		"file1.txt":            &fstest.MapFile{},
		"file2.log":            &fstest.MapFile{},
	}

	tests := []struct {
		name  string
		setup func(*walker.Walker)
		want  string
	}{
		{
			name:  "Walk all files",
			setup: func(w *walker.Walker) {},
			want:  "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; dir2/subdir/file6.go; file1.txt; file2.log",
		},
		{
			name: "Only .txt files",
			setup: func(w *walker.Walker) {
				w.Include(walker.MatchExtension(".txt"))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; file1.txt",
		},
		{
			name: "Exclude .log files",
			setup: func(w *walker.Walker) {
				w.Exclude(walker.MatchExtension(".log"))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Only files in dir1",
			setup: func(w *walker.Walker) {
				w.IncludeDir(walker.MatchRegexpMust("dir1"))
			},
			want: "dir1/file3.txt; dir1/file4.log",
		},
		{
			name: "Exclude dir2",
			setup: func(w *walker.Walker) {
				w.ExcludeDir(walker.MatchRegexpMust("dir2"))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; file1.txt; file2.log",
		},
		{
			name: "Only .txt and .go files",
			setup: func(w *walker.Walker) {
				w.Include(walker.MatchRegexpMust(`\.(txt|go)$`))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Not .go files",
			setup: func(w *walker.Walker) {
				w.Exclude(walker.MatchExtension(".go"))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; file1.txt; file2.log",
		},
		{
			name: "Also not .go files",
			setup: func(w *walker.Walker) {
				w.Include(walker.Not(walker.MatchExtension(".go")))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; file1.txt; file2.log",
		},
		{
			name: "No dot files",
			setup: func(w *walker.Walker) {
				w.Exclude(walker.DotFile)
				w.ExcludeDir(walker.DotFile)
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; dir2/subdir/file6.go; file1.txt; file2.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := walker.New(testFS, ".", walker.HaltOnError)
			tt.setup(&w)

			paths := slices.Collect(w.FilePaths())
			if got := strings.Join(paths, "; "); got != tt.want {
				t.Errorf("Walker.FilePaths: want %q, got %q", tt.want, got)
			}
		})
	}
}

func ExampleWalker() {
	{
		w := walker.New(nil, "testdata", walker.HaltOnError)
		w.Exclude(walker.DotFile)
		w.ExcludeDir(walker.DotFile)
		paths := slices.Collect(w.FilePaths())
		if w.HasError() {
			panic(w.Err())
		}
		fmt.Println("All files:")
		fmt.Println(strings.Join(paths, "; "))
	}
	{
		w := walker.New(nil, "testdata", walker.HaltOnError)
		w.Exclude(walker.DotFile)
		w.ExcludeDir(walker.DotFile)
		w.Include(walker.MatchExtension(".txt"))
		w.IncludeDir(walker.MatchRegexpMust("2"))
		paths := slices.Collect(w.FilePaths())
		if w.HasError() {
			panic(w.Err())
		}
		fmt.Println("Files ending with .txt in a directory with 2:")
		fmt.Println(strings.Join(paths, "; "))
	}
	// Output:
	// All files:
	// testdata/example1/a.txt; testdata/example2/b.txt; testdata/example2/subdir/c.log; testdata/example2/subdir/d.txt
	// Files ending with .txt in a directory with 2:
	// testdata/example2/b.txt; testdata/example2/subdir/d.txt
}

func TestCollectErrors(t *testing.T) {
	dir := t.TempDir()
	testFS := fstest.MapFS{
		"1/a.txt": &fstest.MapFile{},
		"2.txt":   &fstest.MapFile{},
	}
	be.NilErr(t, os.CopyFS(dir, testFS))
	subdir := filepath.Join(dir, "1")
	be.NilErr(t, os.Chmod(subdir, 0o000))
	t.Cleanup(func() {
		be.NilErr(t, os.Chmod(subdir, 0o777))
	})

	var errs []error
	w := walker.New(nil, dir, walker.CollectErrors(&errs))
	paths := slices.Collect(w.FilePaths())
	be.Equal(t, "2.txt", strings.Join(paths, "; "))

	if len(errs) != 1 {
		t.Fatalf("want 1 error; got %v", errs)
	}
	if !errors.Is(errs[0], fs.ErrPermission) {
		t.Errorf("want permission error, got: %v", errs[0])
	}
}
