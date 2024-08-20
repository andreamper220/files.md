package fs

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func init() {
	Ctime = func(fi os.FileInfo) int64 {
		return fi.ModTime().Unix()
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

	r.ElementsMatch([]string{"later", "today", "archive", "-read-", "-shop-", "-watch-", "img", "inbox", "habits", "insights"}, dirNames)
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

func TestFS_TouchNew(t *testing.T) {
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

func TestFS_TouchExisting(t *testing.T) {
	r := require.New(t)
	fs, _ := NewFS("/", afero.NewMemMapFs())
	err := fs.Write("today", "a.md", "A")
	r.NoError(err)

	path := fs.UnsafePath("today", "a.md")
	fi, err := fs.backend.Stat(path)
	r.NoError(err)
	orig_ctime := Ctime(fi)

	time.Sleep(time.Second)
	err = fs.Touch("today", "a.md")
	r.NoError(err)

	fi, err = fs.backend.Stat(path)
	r.NoError(err)
	r.Less(orig_ctime, Ctime(fi))

	content, err := fs.Read("today", "a.md")
	r.NoError(err)
	r.Equal("A", content)
}

func TestFS_GetAllNotesInMatchingDir(t *testing.T) {
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

func TestFS_GetAllMatchingNotesInMatchingDir(t *testing.T) {
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

func TestFS_GetAllNotesInAllMatchingDirs(t *testing.T) {
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

func TestFS_GetAllMatchingNotesInAllMatchingDirs(t *testing.T) {
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

func TestFS_GetAllNotesInAllDirsForEmptyQuery(t *testing.T) {
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

func TestFS_PathTraversalAttack(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	fs.rootPath = "/"

	path := fs.UnsafePath("../root/.ssh/", "authorized_keys")
	r.Equal("/root/.ssh/authorized_keys", path)

	path = fs.UnsafePath("note", "../root/.ssh/authorized_keys")
	r.Equal("/root/.ssh/authorized_keys", path)
}

func TestFS_OnlyUserDirs(t *testing.T) {
	r := require.New(t)

	fs, _ := NewFS("/", afero.NewMemMapFs())
	_ = fs.MakeDir("str")
	_ = fs.MakeDir("123")
	_ = fs.MakeDir("123.56")

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

	fs, _ := NewFS(".", afero.NewMemMapFs())
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
