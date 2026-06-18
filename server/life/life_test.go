package life

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/fs"
)

func init() {
	fs.Ctime = func(fi os.FileInfo) int64 { return 0 }
	fs.Mtime = func(fi os.FileInfo) int64 { return 0 }
}

func newTestFS(t *testing.T) *fs.FS {
	t.Helper()
	fsys, err := fs.NewFS("/", afero.NewMemMapFs())
	require.NoError(t, err)
	return fsys
}

func TestInitCreatesSpheres(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)

	err := Init(fsys)
	r.NoError(err)

	exists, err := fsys.Exists(SpherePath("Работа"), "")
	r.NoError(err)
	r.True(exists)

	exists, err = fsys.Exists(SpherePath("Работа"), SphereHubFile)
	r.NoError(err)
	r.True(exists)
}

func TestCreateProjectStructure(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)
	r.NoError(Init(fsys))

	projectPath, err := CreateProject(fsys, SpherePath("Работа"), "Files MD")
	r.NoError(err)

	for _, sub := range []string{SubDirDrafts, SubDirFinal, SubDirDiscussions} {
		exists, err := fsys.Exists(fs.JoinDir(projectPath, sub), "")
		r.NoError(err)
		r.True(exists)
	}
}

func TestRegisterDocInProject(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)
	r.NoError(Init(fsys))
	projectPath, err := CreateProject(fsys, SpherePath("Работа"), "Test")
	r.NoError(err)

	docDir := DocDir(projectPath, KindDraft)
	r.NoError(RegisterDoc(fsys, KindDraft, docDir, "Idea.md"))

	content, err := fsys.Read(fs.DirUserRoot, IndexFilename)
	r.NoError(err)
	r.Contains(content, "spheres/Работа/Test/drafts/Idea.md")
}

func TestFinalizeDraft(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)
	r.NoError(Init(fsys))
	projectPath, err := CreateProject(fsys, SpherePath("Работа"), "Plan")
	r.NoError(err)

	draftsDir := DocDir(projectPath, KindDraft)
	r.NoError(fsys.Write(draftsDir, "Step.md", "body"))
	r.NoError(RegisterDoc(fsys, KindDraft, draftsDir, "Step.md"))

	err = Finalize(fsys, draftsDir, "Step.md")
	r.NoError(err)

	exists, err := fsys.Exists(DocDir(projectPath, KindFinal), "Step.md")
	r.NoError(err)
	r.True(exists)

	exists, err = fsys.Exists(draftsDir, "Step.md")
	r.NoError(err)
	r.False(exists)
}

func TestMoveProjectToAnotherSphere(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)
	r.NoError(Init(fsys))
	projectPath, err := CreateProject(fsys, SpherePath("Работа"), "Move me")
	r.NoError(err)

	err = MoveEntry(fsys, projectPath, SpherePath("Здоровье"))
	r.NoError(err)

	exists, err := fsys.Exists(projectPath, "")
	r.NoError(err)
	r.False(exists)

	exists, err = fsys.Exists(ProjectPath(SpherePath("Здоровье"), "Move me"), "")
	r.NoError(err)
	r.True(exists)
}

func TestInsertAfterSection(t *testing.T) {
	r := require.New(t)

	content := "## Сферы\n- [A](a.md)\n\n## Проекты\n"
	updated, ok := insertAfterSection(content, "## Сферы", "- [B](b.md)\n")
	r.True(ok)
	r.Contains(updated, "- [B](b.md)")
	r.Contains(updated, "## Проекты")
}
