package server

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/userconfig"
)

func TestReplaceEditNote_ListWithURLs(t *testing.T) {
	r := require.New(t)

	mode := userconfig.DefaultConfig.Mode
	userconfig.DefaultConfig.Mode = userconfig.ModeFull
	defer func() { userconfig.DefaultConfig.Mode = mode }()

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	projectPath := initLifeTestProject(t, userFS)
	finalDir := life.DocDir(projectPath, life.KindFinal)

	const oldContent = "- Oh My Git: https://ohmygit.org/\n\nOh My Git: https://ohmygit.org/\n\nhttps://resu-match.com/\n"
	filename := fs.Filename(fs.SanitizeFilename("- Oh My Git: https://ohmygit.org/"))
	r.NoError(userFS.Write(finalDir, filename, oldContent))

	tgram := tg.NewFakeTG()
	bot, fakeDB := newTestBot(tgram, userFS, fakeConfig())

	dirHash := fs.ShortHash(finalDir)
	fileHash := fs.Hash(filename)

	r.NoError(bot.showFile([]string{dirHash, fileHash}))
	r.NoError(bot.Reply(tg.NewUpdCmd(-1, tg.NewCmd(CmdEditNoteReplace, []string{dirHash, fileHash}))))

	const newContent = "Oh My Git: https://ohmygit.org/\nhttps://resu-match.com/\nГауссовы сплаты: https://vas3k.club/post/27263/\nhttps://www.sqlnoir.com/"
	r.NoError(bot.Reply(tg.NewUpd(-1, newContent)))

	got, err := userFS.Read(finalDir, filename)
	r.NoError(err)
	r.Contains(got, "https://vas3k.club/post/27263/")
	r.Contains(got, "https://www.sqlnoir.com/")
	r.NotContains(got, "resu-match.com/\n\nOh My Git")

	_, _, _, ok := fakeDB.EditNoteTarget()
	r.False(ok, "edit target should be cleared after save")
}

func TestReplaceEditNote_BulletListWithURLs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	projectPath := initLifeTestProject(t, userFS)
	finalDir := life.DocDir(projectPath, life.KindFinal)

	filename := fs.Filename(fs.SanitizeFilename("- Oh My Git: https://ohmygit.org/"))
	r.NoError(userFS.Write(finalDir, filename, "old\n"))

	tgram := tg.NewFakeTG()
	bot, _ := newTestBot(tgram, userFS, fakeConfig())

	dirHash := fs.ShortHash(finalDir)
	fileHash := fs.Hash(filename)

	const newContent = "- Oh My Git: https://ohmygit.org/\n- https://resu-match.com/\n- Гауссовы сплаты: https://vas3k.club/post/27263/\n- https://www.sqlnoir.com/"
	r.NoError(bot.Reply(tg.NewUpdCmd(-1, tg.NewCmd(CmdEditNoteReplace, []string{dirHash, fileHash}))))
	r.NoError(bot.Reply(tg.NewUpd(-1, newContent)))

	got, err := userFS.Read(finalDir, filename)
	r.NoError(err)
	r.Equal(newContent+"\n", got)
}

func TestReplaceEditNote_PreemptsStaleInputExpectation(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	projectPath := initLifeTestProject(t, userFS)
	finalDir := life.DocDir(projectPath, life.KindFinal)

	const oldContent = "old note\n"
	filename := "Resources.md"
	r.NoError(userFS.Write(finalDir, filename, oldContent))

	tgram := tg.NewFakeTG()
	bot, fakeDB := newTestBot(tgram, userFS, fakeConfig())

	dirHash := fs.ShortHash(finalDir)
	fileHash := fs.Hash(filename)

	fakeDB.SetInputExpectation(tg.NewCmd(CmdLifeCreateSection, []string{dirHash, "%s"}))

	r.NoError(bot.Reply(tg.NewUpdCmd(-1, tg.NewCmd(CmdEditNoteReplace, []string{dirHash, fileHash}))))
	r.NoError(bot.Reply(tg.NewUpd(-1, "Oh My Git: https://ohmygit.org/")))

	got, err := userFS.Read(finalDir, filename)
	r.NoError(err)
	r.Equal("Oh My Git: https://ohmygit.org/\n", got)
}
