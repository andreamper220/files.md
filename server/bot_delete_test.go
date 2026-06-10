package server

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/db"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/pkg/tg"
)

func TestDeleteNoteFromShowFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("notes")
	r.NoError(err)
	err = userFS.Write("notes", "My note.md", "Note body")
	r.NoError(err)

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	dirHash := fs.ShortHash("notes")
	fileHash := fs.Hash("My note.md")

	err = bot.showFile([]string{dirHash, fileHash})
	r.NoError(err)

	err = bot.deleteFile([]string{dirHash, fileHash})
	r.NoError(err)

	exists, err := userFS.Exists("notes", "My note.md")
	r.NoError(err)
	r.False(exists)
}

func TestDeleteDirFromShowDirs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("notes")
	r.NoError(err)
	err = userFS.Write("notes", "Inside.md", "content")
	r.NoError(err)

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	dirHash := fs.ShortHash("notes")
	err = bot.deleteDir([]string{dirHash})
	r.NoError(err)

	exists, err := userFS.Exists("notes", "")
	r.NoError(err)
	r.False(exists)
}

func TestCannotDeleteProtectedRootFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write(fs.DirUserRoot, fs.ChatFilename, "inbox")
	r.NoError(err)

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	r.False(canDeleteNote(fs.DirUserRoot, fs.ChatFilename))

	err = bot.deleteFile([]string{"", fs.Hash(fs.ChatFilename)})
	r.Error(err)
}

func TestCannotDeleteSystemDir(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateSystemDirs()
	r.NoError(err)

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	r.False(canDeleteDir(fs.DirJournal))

	err = bot.deleteDir([]string{fs.ShortHash(fs.DirJournal)})
	r.Error(err)
}
