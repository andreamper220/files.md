package server

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/userconfig"
)

func TestResolveMediaPath_ByFullPath(t *testing.T) {
	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.Write(fs.DirMedia, "go-cheatsheet-1page.pdf", "pdf"))

	bot, _ := newTestBot(tg.NewFakeTG(), userFS, userconfig.NewConfig(userFS, -1, "config.yaml"))

	got, err := bot.resolveMediaPath("media/go-cheatsheet-1page.pdf")
	r.NoError(err)
	r.Equal("media/go-cheatsheet-1page.pdf", got)
}

func TestResolveMediaPath_ByBasename(t *testing.T) {
	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.Write(fs.DirMedia, "go-cheatsheet-1page.pdf", "pdf"))

	bot, _ := newTestBot(tg.NewFakeTG(), userFS, userconfig.NewConfig(userFS, -1, "config.yaml"))

	got, err := bot.resolveMediaPath("go-cheatsheet-1page.pdf")
	r.NoError(err)
	r.Equal("media/go-cheatsheet-1page.pdf", got)
}

func TestResolveMediaPath_ByHash(t *testing.T) {
	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.Write(fs.DirMedia, "report.pdf", "pdf"))

	bot, _ := newTestBot(tg.NewFakeTG(), userFS, userconfig.NewConfig(userFS, -1, "config.yaml"))
	path := "media/report.pdf"

	got, err := bot.resolveMediaPath(fs.Hash(path))
	r.NoError(err)
	r.Equal(path, got)
}

func TestMediaOpenCmd_PrefersPathOverShortHash(t *testing.T) {
	r := require.New(t)
	path := "media/go-cheatsheet-1page.pdf"
	cmd := mediaOpenCmd(path, "go-cheatsheet-1page.pdf")
	r.Equal(CmdOpenMedia, cmd.Name)
	r.Equal([]string{path}, cmd.Params)
}
