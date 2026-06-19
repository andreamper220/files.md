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
	r.Contains(report, "💼")
}
