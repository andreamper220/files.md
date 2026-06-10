package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/pkg/tg"
)

const (
	CmdShowDeleteFile = "del_f"
	CmdDeleteFile     = "rm_f"
	CmdShowDeleteDir  = "del_d"
	CmdDeleteDir      = "rm_d"
)

var protectedRootFiles = []string{
	fs.ChatFilename,
	fs.LaterFilename,
	fs.DoneFilename,
	fs.ShopFilename,
	fs.WatchFilename,
	fs.ReadFilename,
}

func canDeleteNote(dir, filename string) bool {
	if filepath.Ext(filename) != fs.MDExt {
		return false
	}

	for _, part := range pathParts(dir) {
		if isProtectedDirName(part) {
			return false
		}
	}

	if dir == fs.DirUserRoot || dir == "" {
		for _, protected := range protectedRootFiles {
			if filename == protected {
				return false
			}
		}
	}

	base := strings.TrimSuffix(filename, fs.MDExt)
	if strings.HasPrefix(base, "_") && strings.HasSuffix(base, "_") {
		return false
	}

	return true
}

func canDeleteDir(dir string) bool {
	dir = strings.Trim(dir, "/")
	if dir == "" {
		return false
	}

	for _, part := range pathParts(dir) {
		if isProtectedDirName(part) {
			return false
		}
	}

	return true
}

func isProtectedDirName(name string) bool {
	return strings.EqualFold(name, fs.DirMedia) ||
		strings.EqualFold(name, fs.DirArchive) ||
		strings.EqualFold(name, fs.DirJournal) ||
		strings.EqualFold(name, fs.DirHabits) ||
		strings.EqualFold(name, fs.DirInsights) ||
		(strings.HasPrefix(name, "_") && strings.HasSuffix(name, "_"))
}

func pathParts(dir string) []string {
	dir = strings.Trim(dir, "/")
	if dir == "" || dir == fs.DirUserRoot {
		return nil
	}
	return strings.Split(dir, "/")
}

func dirHashParam(dir string) string {
	if dir == fs.DirUserRoot || dir == "" {
		return ""
	}
	return fs.ShortHash(dir)
}

func (b *Bot) showDeleteFileConfirm(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, filename, err := b.resolveNotePath(dirHash, filenameHash)
	if err != nil {
		return err
	}
	if !canDeleteNote(dir, filename) {
		return fmt.Errorf("delete file: protected note %s/%s", dir, filename)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrBack), tg.NewCmd(CmdShowFile, []string{dirHash, filenameHash})),
			tg.NewBtn(i18n.Tr("Yes, delete"), tg.NewCmd(CmdDeleteFile, []string{dirHash, filenameHash})),
		),
	})

	msg := b.tr("Delete <b>%s</b>? This cannot be undone.", fs.DisplayName(filename))
	return b.showHTML(msg, kb)
}

func (b *Bot) deleteFile(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, filename, err := b.resolveNotePath(dirHash, filenameHash)
	if err != nil {
		return err
	}
	if !canDeleteNote(dir, filename) {
		return fmt.Errorf("delete file: protected note %s/%s", dir, filename)
	}

	if err := b.fs.Del(dir, filename); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	b.delAllKeyboards()
	_, _ = b.tg.Send(b.userID, b.tr("Deleted <b>%s</b>", fs.DisplayName(filename)), nil, tg.MarkupHTML)

	return b.showDirs([]string{dirHashParam(dir)})
}

func (b *Bot) showDeleteDirConfirm(params []string) error {
	dir, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("delete dir confirm: %w", err)
	}
	if !canDeleteDir(dir) {
		return fmt.Errorf("delete dir confirm: protected dir %s", dir)
	}

	dirHash := fs.ShortHash(dir)
	parent := fs.ParentDirPath(dir)
	backParam := dirHashParam(parent)

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrBack), tg.NewCmd(CmdShowDirs, []string{backParam})),
			tg.NewBtn(i18n.Tr("Yes, delete"), tg.NewCmd(CmdDeleteDir, []string{dirHash})),
		),
	})

	msg := b.tr("Delete folder <b>%s</b> and all its contents? This cannot be undone.", fs.DisplayName(dir))
	return b.showHTML(msg, kb)
}

func (b *Bot) deleteDir(params []string) error {
	dir, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("delete dir: %w", err)
	}
	if !canDeleteDir(dir) {
		return fmt.Errorf("delete dir: protected dir %s", dir)
	}

	if err := b.fs.DelDir(dir); err != nil {
		return fmt.Errorf("delete dir: %w", err)
	}

	b.delAllKeyboards()
	_, _ = b.tg.Send(b.userID, b.tr("Deleted folder <b>%s</b>", fs.DisplayName(dir)), nil, tg.MarkupHTML)

	return b.showDirs([]string{dirHashParam(fs.ParentDirPath(dir))})
}

func (b *Bot) resolveNotePath(dirHash, filenameHash string) (string, string, error) {
	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return "", "", fmt.Errorf("resolve note path: can't find dir: %w", err)
	}

	var filename string
	if strings.Contains(filenameHash, "/") {
		filename = filenameHash
	} else {
		filename, err = b.fs.Unhash(dir, filenameHash)
		if err != nil {
			return "", "", fmt.Errorf("resolve note path: can't find file: %w", err)
		}
	}

	return dir, filename, nil
}
