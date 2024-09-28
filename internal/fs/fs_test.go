package fs

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func init() {
	Ctime = func(fi os.FileInfo) int64 {
		return 0
	}
}

func TestIsChecklistItem(t *testing.T) {
	r := require.New(t)

	r.False(IsChecklistItem("-checklist-"))
	r.True(IsChecklistItem("-checklist-item"))
}

func TestTitle(t *testing.T) {
	r := require.New(t)

	title := Title("filename")
	r.Equal("Filename", title)
}

func TestTitleWithSpace(t *testing.T) {
	r := require.New(t)

	title := Title(" filename ")
	r.Equal("Filename", title)
}

func TestTitleChecklist(t *testing.T) {
	r := require.New(t)

	title := Title("-checklist-")
	r.Equal("Checklist", title)
}

func TestTitleChecklistItem(t *testing.T) {
	r := require.New(t)

	title := Title("-checklist-item")
	r.Equal("Item", title)
}

func TestMD5(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	res := fs.md5("First task.md")

	r.Equal("0824149b387", res)
}

func TestExcludeChecklists(t *testing.T) {
	r := require.New(t)

	noChecklists := ExcludeChecklists([]File{{Name: "not-a-checklist"}, {Name: "-checklist-"}})

	r.Equal([]File{{Name: "not-a-checklist"}}, noChecklists)
}

func TestExcludeSystemDirs(t *testing.T) {
	r := require.New(t)

	noChecklists := ExcludeSystemDirs([]File{{Name: "not-a-system-dir"}, {Name: "img"}, {Name: "archive"}, {Name: "journal"}})

	r.Equal([]File{{Name: "not-a-system-dir"}}, noChecklists)
}

func TestExcludeTaskDirs(t *testing.T) {
	r := require.New(t)

	noChecklists := ExcludeTaskDirs([]File{{Name: "not-a-task-dir"}, {Name: "today"}, {Name: "later"}})

	r.Equal([]File{{Name: "not-a-task-dir"}}, noChecklists)
}

func TestIsMultiline(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "First task.md", "")
	r.NoError(err)

	isMultiline, err := fs.IsMultiline("today", "First task.md")
	r.NoError(err)
	r.False(isMultiline)

	err = fs.Write("today", "Second task.md", "c")
	r.NoError(err)

	isMultiline, err = fs.IsMultiline("today", "Second task.md")
	r.NoError(err)
	r.True(isMultiline)

	err = fs.Write("today", "Third task.md", " \n ")
	r.NoError(err)

	isMultiline, err = fs.IsMultiline("today", "Third task.md")
	r.NoError(err)
	r.False(isMultiline)
}

func TestGetFilesInDir(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "First task.md", "")
	r.NoError(err)

	files, err := fs.FilesAndDirs("today")
	r.NoError(err)
	r.Len(files, 1)
	r.Equal("First task.md", files[0].Name)
}

func TestCreateBaseDirs(t *testing.T) {
	r := require.New(t)

	fs, err := NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(fs.CreateDirsIfNotExist())

	err = fs.CreateDirsIfNotExist()
	r.NoError(err)

	dirs, err := fs.FilesAndDirs("")
	r.NoError(err)
	dirs = OnlyDirs(dirs)
	dirNames := OnlyFilenames(dirs)

	r.ElementsMatch([]string{"later", "today", "archive", "-read-", "-shop-", "-watch-", "img", "habits", "journal", "insights"}, dirNames)
}

func TestSortByCtimeDesc(t *testing.T) {
	r := require.New(t)

	saved := Ctime
	defer func() {
		Ctime = saved
	}()
	Ctime = func(fi os.FileInfo) int64 {
		if fi.Name() == "b.md" {
			return 1
		}

		return 2
	}

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "b.md", "")
	r.NoError(err)

	err = fs.Write("today", "a.md", "")
	r.NoError(err)

	entries, err := fs.FilesAndDirs("today")
	r.NoError(err)

	r.Equal([]string{"a.md", "b.md"}, OnlyFilenames(SortByCtimeDesc(entries)))
}

func TestExcludeEverythingButUserDirs(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("", "a.md", "")
	r.NoError(err)

	err = fs.MakeDir("dir")
	r.NoError(err)

	entries, err := fs.FilesAndDirs("")
	r.NoError(err)

	dirs := OnlyDirs(ExcludeTaskDirs(ExcludeSystemDirs(entries)))
	r.Len(dirs, 1)
	r.Equal("dir", dirs[0].Name)
}

func TestOnlyFiles(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("", "a.md", "")
	r.NoError(err)

	err = fs.MakeDir("dir")
	r.NoError(err)

	entries, err := fs.FilesAndDirs("")
	r.NoError(err)

	dirs := OnlyMDFiles(entries)
	r.Len(dirs, 1)
	r.Equal("a.md", dirs[0].Name)
}

func TestOnlyChecklists(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "a.md", "")
	r.NoError(err)

	err = fs.MakeDir("-list-")
	r.NoError(err)

	entries, err := fs.FilesAndDirs("")
	r.NoError(err)

	dirs := OnlyChecklists(entries)
	r.Len(dirs, 1)
	r.Equal("-list-", dirs[0].Name)
}

func TestFSTouchNew(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	exists, err := fs.Exists("today", "a.md")
	r.NoError(err)
	r.False(exists)

	err = fs.Touch("today", "a.md")
	r.NoError(err)

	exists, err = fs.Exists("today", "a.md")
	r.NoError(err)
	r.True(exists)
}

func TestFSTouchExisting(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "a.md", "A")
	r.NoError(err)

	err = fs.Touch("today", "a.md")
	r.NoError(err)

	content, err := fs.Read("today", "a.md")
	r.NoError(err)
	r.Equal("A", content)
}

func TestFSGetAllNotesInMatchingDir(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("brain", "a.md")
	r.NoError(err)
	err = fs.Touch("today", "b.md")
	r.NoError(err)
	err = fs.Touch("non-matching-dir", "c.md")
	r.NoError(err)

	notes, err := fs.SearchNotes("BRAIN")
	r.NoError(err)
	r.Len(notes, 1)
	r.Equal("a.md", notes[0].Name)
}

func TestFSGetAllMatchingNotesInMatchingDir(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("brain", "a.md")

	r.NoError(err)
	err = fs.Touch("brain", "b.md")
	r.NoError(err)
	err = fs.Touch("today", "c.md")
	r.NoError(err)

	notes, err := fs.SearchNotes("BRAIN A")
	r.NoError(err)
	r.Len(notes, 1)
	r.Equal("a.md", notes[0].Name)
}

func TestFSGetAllNotesInAllMatchingDirs(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("brain", "a.md")
	r.NoError(err)
	err = fs.Touch("brain", "b.md")
	r.NoError(err)
	err = fs.Touch("today", "c.md")
	r.NoError(err)

	notes, err := fs.SearchNotes("brain")
	r.NoError(err)
	r.Len(notes, 2)

	var noteFilenames []string
	for _, note := range notes {
		noteFilenames = append(noteFilenames, note.Name)
	}

	r.ElementsMatch([]string{"a.md", "b.md"}, noteFilenames)
}

func TestFSGetAllMatchingNotesInAllMatchingDirs(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("brain", "a.md")
	r.NoError(err)
	err = fs.Touch("brain", "ab.md")
	r.NoError(err)
	err = fs.Touch("brain", "b.md")
	r.NoError(err)
	err = fs.Touch("today", "c.md")
	r.NoError(err)

	notes, err := fs.SearchNotes("brain a")
	r.NoError(err)
	r.Len(notes, 2)

	var noteFilenames []string
	for _, note := range notes {
		noteFilenames = append(noteFilenames, note.Name)
	}

	r.ElementsMatch([]string{"a.md", "ab.md"}, noteFilenames)
}

func TestFSGetAllNotesInAllDirsForEmptyQuery(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("brain", "a.md")
	r.NoError(err)
	err = fs.Touch("b", "b.md")
	r.NoError(err)
	err = fs.Touch("today", "c.md")
	r.NoError(err)

	notes, err := fs.SearchNotes("")
	r.NoError(err)
	r.Len(notes, 2)

	var noteFilenames []string
	for _, note := range notes {
		noteFilenames = append(noteFilenames, note.Name)
	}

	r.ElementsMatch([]string{"a.md", "b.md"}, noteFilenames)
}

func TestFSPathTraversalAttack(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	fs.RootPath = "/"

	path := fs.UnsafePath("../root/.ssh/", "authorized_keys")
	r.Equal("/root/.ssh/authorized_keys", path)

	path = fs.UnsafePath("note", "../root/.ssh/authorized_keys")
	r.Equal("/root/.ssh/authorized_keys", path)
}

func TestFSOnlyUserDirs(t *testing.T) {
	r := require.New(t)

	fs, err := NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	err = fs.MakeDir("str")
	r.NoError(err)

	err = fs.MakeDir("123")
	r.NoError(err)

	err = fs.MakeDir("123.56")
	r.NoError(err)

	dirs, _ := fs.FilesAndDirs("")
	userDirs := OnlyUserDirs(dirs)

	r.Len(userDirs, 1)
	r.Equal("123", userDirs[0].Name)
}

func TestIsSafeWrongRoot(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/a", afero.NewMemMapFs())
	r.False(fs.isSafe("/b"))
}

func TestIsSafePathTraversalAttack(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/a", afero.NewMemMapFs())
	r.False(fs.isSafe("/a/../b"))
	r.False(fs.isSafe("/a/../../b"))
	r.False(fs.isSafe("./a/../b"))
	r.False(fs.isSafe("./a/../../b"))
}

func TestIsSafePathTraversalAttackWithRelativePaths(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS(".", afero.NewMemMapFs())
	r.False(fs.isSafe("./a/../b"))
	r.False(fs.isSafe("./a/../../b"))
}

func TestUnhashRootDirectory(t *testing.T) {
	r := require.New(t)

	fs, err := NewFS(".", afero.NewMemMapFs())
	r.NoError(err)
	unhashed, err := fs.Unhash("", "")
	r.NoError(err)

	r.Equal("", unhashed)
}

func TestSanitizeFilename(t *testing.T) {
	r := require.New(t)

	r.Equal("ab", SanitizeFilename("a\x00b"))
	r.Equal("a{|}b", SanitizeFilename("a/b"))
	r.Equal("a{||}b", SanitizeFilename("a\\b"))
	r.Equal("a{|}b{||}", SanitizeFilename("\x00a\x00/b\\"))
}

func TestUnsanitizeFilename(t *testing.T) {
	r := require.New(t)

	r.Equal("a/b", UnsanitizeFilename("a{|}b"))
	r.Equal("a\\b", UnsanitizeFilename("a{||}b"))
	r.Equal("a/b\\", UnsanitizeFilename("a{|}b{||}"))
}

func TestExists(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "First task.md", "")
	r.NoError(err)

	exists, err := fs.Exists("today", "First task.md")
	r.NoError(err)
	r.True(exists)
}

func TestExistsRoot(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "First task.md", "")
	r.NoError(err)

	exists, err := fs.Exists("", "")
	r.NoError(err)
	r.True(exists)
}

func TestWriteAndReadFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "test.md", "Test content")
	r.NoError(err)

	content, err := fs.Read("today", "test.md")
	r.NoError(err)
	r.Equal("Test content", content)
}

func TestWriteUnsafePath(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/1", afero.NewMemMapFs())
	err := fs.Write("../unsafe", "test.md", "Test content")
	r.Error(err)
	r.Contains(err.Error(), "unsafe filePath")
}

func TestReadNonExistentFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	content, err := fs.Read("today", "nonexistent.md")
	r.Error(err)
	r.Equal("", content)
}

func TestDelFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "delete.md", "To be deleted")
	r.NoError(err)

	err = fs.Del("today", "delete.md")
	r.NoError(err)

	exists, err := fs.Exists("today", "delete.md")
	r.NoError(err)
	r.False(exists)
}

func TestDelNonExistentFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Del("today", "nonexistent.md")
	r.Error(err)
	r.Contains(err.Error(), "can't remove")
}

func TestRenameFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "oldname.md", "Old content")
	r.NoError(err)

	err = fs.Rename("today", "oldname.md", "today", "newname.md")
	r.NoError(err)

	exists, err := fs.Exists("today", "newname.md")
	r.NoError(err)
	r.True(exists)

	content, err := fs.Read("today", "newname.md")
	r.NoError(err)
	r.Equal("Old content", content)
}

func TestRenameUnsafePath(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/1", afero.NewMemMapFs())
	err := fs.Write("today", "oldname.md", "Old content")
	r.NoError(err)

	err = fs.Rename("../unsafe", "oldname.md", "today", "newname.md")
	r.Error(err)
	r.Contains(err.Error(), "unsafe path")
}

func TestMakeDir(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.MakeDir("newdir")
	r.NoError(err)

	exists, err := fs.Exists("newdir", "")
	r.NoError(err)
	r.True(exists)
}

func TestMakeDirUnsafePath(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/1", afero.NewMemMapFs())
	err := fs.MakeDir("../unsafe")
	r.Error(err)
	r.Contains(err.Error(), "unsafe path")
}

func TestUnsafePathSanitization(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/1", afero.NewMemMapFs())

	unsafePath := fs.UnsafePath("../", "test.md")
	r.Equal("/test.md", unsafePath)

	unsafePath = fs.UnsafePath("safe", "../unsafe.md")
	r.Equal("/1/unsafe.md", unsafePath)
}

func TestUnhash(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("today", "hashedfile.md")
	r.NoError(err)

	hash := fs.md5("hashedfile.md")
	unhashed, err := fs.Unhash("today", hash)
	r.NoError(err)
	r.Equal("hashedfile.md", unhashed)
}

func TestUnhashNonExistentFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	_, err := fs.Unhash("today", "nonexistenthash")
	r.Error(err)
	r.Contains(err.Error(), "can't unhash:")
}

func TestSanitizeAndUnsanitizeFilename(t *testing.T) {
	r := require.New(t)

	sanitized := SanitizeFilename("test/file:name\\with/special\\chars")
	r.Equal("test{|}file꞉name{||}with{|}special{||}chars", sanitized)

	unsanitized := UnsanitizeFilename(sanitized)
	r.Equal("test/file꞉name\\with/special\\chars", unsanitized)
}

func TestFilesAndDirs(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.MakeDir("dir1")
	r.NoError(err)

	err = fs.Write("dir1", "file1.md", "File content")
	r.NoError(err)

	files, err := fs.FilesAndDirs("dir1")
	r.NoError(err)
	r.Len(files, 1)
	r.Equal("file1.md", files[0].Name)
}

func TestDirs(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.MakeDir("dir1")
	r.NoError(err)

	err = fs.Write("dir1", "file1.md", "File content")
	r.NoError(err)

	dirs, err := fs.Dirs()
	r.NoError(err)
	r.Len(dirs, 1)
	r.Equal("dir1", dirs[0].Name)
}

func TestTouchFile(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Touch("today", "touchfile.md")
	r.NoError(err)

	exists, err := fs.Exists("today", "touchfile.md")
	r.NoError(err)
	r.True(exists)
}
