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

func TestRanger(t *testing.T) {
	testFS := fstest.MapFS{
		"a.txt":                &fstest.MapFile{},
		"dir1/file3.txt":       &fstest.MapFile{},
		"dir1/file4.log":       &fstest.MapFile{},
		"dir2/file5.txt":       &fstest.MapFile{},
		"dir2/subdir/file6.go": &fstest.MapFile{},
		"file1.txt":            &fstest.MapFile{},
		"file2.log":            &fstest.MapFile{},
	}

	temp := t.TempDir()
	be.NilErr(t, os.CopyFS(temp, testFS))

	tests := []struct {
		name  string
		setup func(*walker.Ranger)
		want  string
	}{
		{
			name:  "Walk all files",
			setup: func(w *walker.Ranger) {},
			want:  "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; dir2/subdir/file6.go; file1.txt; file2.log",
		},
		{
			name: "Only .txt files",
			setup: func(w *walker.Ranger) {
				w.Include(walker.MatchExtension(".txt"))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; file1.txt",
		},
		{
			name: "Exclude .log files",
			setup: func(w *walker.Ranger) {
				w.Exclude(walker.MatchExtension(".log"))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Only files in dir1",
			setup: func(w *walker.Ranger) {
				w.IncludeDir(walker.MatchRegexpMust("dir1"))
			},
			want: "dir1/file3.txt; dir1/file4.log",
		},
		{
			name: "Exclude dir2",
			setup: func(w *walker.Ranger) {
				w.ExcludeDir(walker.MatchRegexpMust("dir2"))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; file1.txt; file2.log",
		},
		{
			name: "Only .txt and .go files",
			setup: func(w *walker.Ranger) {
				w.Include(walker.MatchRegexpMust(`\.(txt|go)$`))
			},
			want: "a.txt; dir1/file3.txt; dir2/file5.txt; dir2/subdir/file6.go; file1.txt",
		},
		{
			name: "Not .go files",
			setup: func(w *walker.Ranger) {
				w.Exclude(walker.MatchExtension(".go"))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; file1.txt; file2.log",
		},
		{
			name: "Also not .go files",
			setup: func(w *walker.Ranger) {
				w.Include(walker.Not(walker.MatchExtension(".go")))
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; file1.txt; file2.log",
		},
		{
			name: "No dot files",
			setup: func(w *walker.Ranger) {
				w.Exclude(walker.MatchDotFile)
				w.ExcludeDir(walker.MatchDotFile)
			},
			want: "a.txt; dir1/file3.txt; dir1/file4.log; dir2/file5.txt; dir2/subdir/file6.go; file1.txt; file2.log",
		},
		{
			name: "Files in dir*",
			setup: func(w *walker.Ranger) {
				w.IncludeDir(walker.MatchGlobName("dir*"))
			},
			want: "dir1/file3.txt; dir1/file4.log; dir2/file5.txt",
		},
		{
			name: "Log files in dir*",
			setup: func(w *walker.Ranger) {
				w.Include(walker.MatchExtension(".log"))
				w.IncludeDir(walker.MatchGlobName("dir*"))
			},
			want: "dir1/file4.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := walker.New(testFS, ".", walker.OnErrorHalt)
			tt.setup(&tr)

			paths := slices.Collect(tr.FilePaths())
			be.Equal(t, tt.want, strings.Join(paths, "; "))

			tr = walker.New(nil, temp, walker.OnErrorHalt)
			tt.setup(&tr)

			paths = nil
			prefix := temp + string(filepath.Separator)
			for path := range tr.FilePaths() {
				paths = append(paths, strings.TrimPrefix(path, prefix))
			}
			be.Equal(t, tt.want, strings.Join(paths, "; "))

			paths = nil
			for entry := range tr.FileEntries() {
				paths = append(paths, strings.TrimPrefix(entry.Path, prefix))
			}
			be.Equal(t, tt.want, strings.Join(paths, "; "))

			paths = nil
			for path, _ := range tr.Files() {
				paths = append(paths, strings.TrimPrefix(path, prefix))
			}
			be.Equal(t, tt.want, strings.Join(paths, "; "))
		})
	}
}

func TestRanger_iter_break(t *testing.T) {
	testFS := fstest.MapFS{
		"a.txt":                &fstest.MapFile{},
		"dir1/file3.txt":       &fstest.MapFile{},
		"dir1/file4.log":       &fstest.MapFile{},
		"dir2/file5.txt":       &fstest.MapFile{},
		"dir2/subdir/file6.go": &fstest.MapFile{},
		"file1.txt":            &fstest.MapFile{},
		"file2.log":            &fstest.MapFile{},
	}

	tr := walker.New(testFS, ".", walker.OnErrorHalt)
	for range tr.Entries() {
		break
	}
	for range tr.FilesAndDirs() {
		break
	}
	for range tr.FileEntries() {
		break
	}
	for range tr.Files() {
		break
	}
	for range tr.FilePaths() {
		break
	}
}

func ExampleRanger() {
	// Demo filesystem
	fsys := fstest.MapFS{
		".a-stuff/file-1.jpeg": &fstest.MapFile{},
		"b-file-2.txt":         &fstest.MapFile{},
		"c/file-3.png":         &fstest.MapFile{},
		"d/file-4.jpeg":        &fstest.MapFile{},
		"e-file-5.txt":         &fstest.MapFile{},
	}

	// Make a new Ranger that halts on error
	tr := walker.New(fsys, ".", walker.OnErrorHalt)

	fmt.Println("Files:")
	for path := range tr.FilePaths() {
		fmt.Println("-", path)
	}
	// Do a final error check
	if tr.HasError() {
		panic(tr.Err())
	}
	// Output:
	// Files:
	// - .a-stuff/file-1.jpeg
	// - b-file-2.txt
	// - c/file-3.png
	// - d/file-4.jpeg
	// - e-file-5.txt
}

func ExampleRanger_matching() {
	// Demo filesystem
	fsys := fstest.MapFS{
		".a-stuff/file-1.jpeg": &fstest.MapFile{},
		"b-file-2.txt":         &fstest.MapFile{},
		"c/file-3.png":         &fstest.MapFile{},
		"d/file-4.jpeg":        &fstest.MapFile{},
		"e-file-5.txt":         &fstest.MapFile{},
	}

	// Make a new Ranger that ignores permission errors and halts for other problems
	tr := walker.New(fsys, ".", walker.OnErrPermissionIgnore)
	// Ignore dot files and dot directories
	tr.Exclude(walker.MatchDotFile)
	tr.ExcludeDir(walker.MatchDotFile)
	// Only list PNG and JPEG files
	tr.Include(walker.Or(
		walker.MatchExtension(".png"),
		walker.MatchExtension(".jpeg"),
	))

	fmt.Println("Files:")
	for path := range tr.FilePaths() {
		fmt.Println("-", path)
	}
	// Do a final error check
	if tr.HasError() {
		panic(tr.Err())
	}
	// Output:
	// Files:
	// - c/file-3.png
	// - d/file-4.jpeg
}

func tempDirWithPermErr(t *testing.T) string {
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
	return dir
}

func TestOnErrorHalt(t *testing.T) {
	dir := tempDirWithPermErr(t)

	w := walker.New(nil, dir, walker.OnErrorHalt)
	var paths []string
	for path := range w.FilePaths() {
		paths = append(paths, filepath.Base(path))
	}
	be.Equal(t, "", strings.Join(paths, "; "))
	be.True(t, errors.Is(w.Err(), fs.ErrPermission))
}

func TestOnErrPermissionIgnore(t *testing.T) {
	dir := tempDirWithPermErr(t)

	w := walker.New(nil, dir, walker.OnErrPermissionIgnore)
	var paths []string
	for path := range w.FilePaths() {
		paths = append(paths, filepath.Base(path))
	}
	be.Equal(t, "2.txt", strings.Join(paths, "; "))
	be.NilErr(t, w.Err())
}

func TestCollectErrors(t *testing.T) {
	dir := tempDirWithPermErr(t)

	var errs []error
	w := walker.New(nil, dir, walker.OnErrorCollect(&errs))
	var paths []string
	for path := range w.FilePaths() {
		paths = append(paths, filepath.Base(path))
	}
	be.Equal(t, "2.txt", strings.Join(paths, "; "))
	be.Equal(t, 1, len(errs))
	be.True(t, errors.Is(errs[0], fs.ErrPermission))
}
