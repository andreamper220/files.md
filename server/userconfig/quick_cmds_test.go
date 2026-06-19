package userconfig

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/fs"
)

func TestAddAndDelQuickCmd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	cfg := NewConfig(userFS, -1, "config.json")
	err = cfg.CreateDefaultIfNotExists()
	r.NoError(err)

	err = cfg.AddQuickCmd("new_quick_cmd")
	r.NoError(err)

	cmds, err := cfg.QuickCmds()
	r.NoError(err)
	r.Equal([]string{"new_quick_cmd"}, cmds)

	c, err := userFS.Read("", "config.json")
	r.NoError(err)
	r.Equal(defaultConfigJSON(t, func(c *config) {
		c.QuickCmds = []string{"new_quick_cmd"}
	}), c)

	err = cfg.DelQuickCmd("new_quick_cmd")
	r.NoError(err)

	cmds, err = cfg.QuickCmds()
	r.NoError(err)
	r.Empty(cmds)

	c, err = userFS.Read("", "config.json")
	r.NoError(err)
	r.Equal(defaultConfigJSON(t, nil), c)
}
