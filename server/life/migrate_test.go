package life

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/fs"
)

func TestMigrateFromFolders_NestedDirsToFinal(t *testing.T) {
	r := require.New(t)
	fsys, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(Init(fsys))

	r.NoError(fsys.Write("resources", "Oh My Git.md", "https://ohmygit.org/\n"))
	r.NoError(fsys.MakeDir("resources/links"))
	r.NoError(fsys.Write("resources/links", "SQL Noir.md", "https://www.sqlnoir.com/\n"))

	res, err := MigrateFromFolders(fsys, MigrateOptions{
		Sphere: "Личное",
		Kind:   KindFinal,
		Dirs:   []string{"resources"},
	})
	r.NoError(err)
	r.Equal(1, res.DirsProcessed)
	r.Equal(1, res.ProjectsCreated)
	r.Equal(1, res.SectionsCreated)
	r.Equal(2, res.FilesMoved)

	_, err = fsys.Read("spheres/Личное/resources/final", "Oh My Git.md")
	r.NoError(err)
	_, err = fsys.Read("spheres/Личное/resources/links/final", "SQL Noir.md")
	r.NoError(err)

	exists, err := fsys.Exists("resources", "")
	r.NoError(err)
	r.False(exists)
}

func TestMigrateFromFolders_RootMD(t *testing.T) {
	r := require.New(t)
	fsys, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(Init(fsys))

	r.NoError(fsys.Write(fs.DirUserRoot, "Quick idea.md", "body\n"))

	res, err := MigrateFromFolders(fsys, MigrateOptions{
		Sphere:      "Работа",
		Kind:        KindFinal,
		RootMD:      true,
		RootProject: "Inbox",
	})
	r.NoError(err)
	r.Equal(1, res.ProjectsCreated)
	r.Equal(1, res.FilesMoved)

	_, err = fsys.Read("spheres/Работа/Inbox/final", "Quick idea.md")
	r.NoError(err)
}

func TestSphereForLegacyDir(t *testing.T) {
	if SphereForLegacyDir("work-notes") != "Работа" {
		t.Fatal("expected Работа")
	}
	if SphereForLegacyDir("brain") != "Личное" {
		t.Fatal("expected Личное")
	}
}
