package morningsummary

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/userconfig"
)

func TestBuild_IncludesNotesSection(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	savedMtime := fs.Mtime
	defer func() {
		fs.Ctime = savedCtime
		fs.Mtime = savedMtime
	}()

	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	fs.Ctime = func(fi os.FileInfo) int64 { return now.UnixMicro() }
	fs.Mtime = func(fi os.FileInfo) int64 { return now.UnixMicro() }

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateSystemDirs()
	r.NoError(err)
	r.NoError(life.Init(userFS))

	cfg := userconfig.NewConfig(userFS, 1, "config.json")
	r.NoError(cfg.CreateDefaultIfNotExists())

	report, err := Build(userFS, cfg)
	r.NoError(err)
	r.Contains(report, "💼 Работа")
	r.Contains(report, "🔴0")
	r.Contains(report, "📝 0")
	r.NotContains(report, "✨")
	r.NotContains(report, "💬")
}

func TestBuild_ShowsOpenTaskCountsByPriority(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	defer func() { fs.Ctime = savedCtime }()
	fs.Ctime = func(fi os.FileInfo) int64 { return 0 }

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.CreateSystemDirs())
	r.NoError(life.Init(userFS))

	spheres, err := life.ListSpheres(userFS)
	r.NoError(err)
	r.NotEmpty(spheres)

	projectPath, err := life.CreateProject(userFS, spheres[0], "Тест")
	r.NoError(err)

	err = userFS.Write(projectPath, life.TasksFilename, "- [ ] 🔴 Open one\n- [ ] 🟠 Open two\n")
	r.NoError(err)

	cfg := userconfig.NewConfig(userFS, 1, "config.json")
	r.NoError(cfg.CreateDefaultIfNotExists())

	report, err := Build(userFS, cfg)
	r.NoError(err)
	r.Contains(report, "🔴1")
	r.Contains(report, "🟠1")
	r.Contains(report, "Тест")
	r.NotContains(report, "⚪️")
}
