package server

import (
	"fmt"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
)

func (b *Bot) showEditNote(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show edit note: missing params")
	}
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("show edit note: %w", err)
	}
	if _, err := b.fs.Unhash(dir, filenameHash); err != nil {
		return fmt.Errorf("show edit note: %w", err)
	}

	back := tg.NewCmd(CmdShowFile, []string{dirHash, filenameHash})
	kb := tg.NewKeyboard([]tg.Row{tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back))})

	cmd := tg.NewCmd(CmdEditNote, []string{dirHash, filenameHash, "%s"})
	b.db.SetInputExpectation(cmd)

	return b.showHTML(i18n.Tr("OK. Send me the new text for your note"), kb)
}

func (b *Bot) editNote(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("edit note: missing params")
	}
	dirHash := params[0]
	filenameHash := params[1]
	newText := strings.TrimSpace(params[2])
	if newText == "" {
		return fmt.Errorf("edit note: empty text")
	}

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("edit note: %w", err)
	}
	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("edit note: %w", err)
	}

	existing, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("edit note: %w", err)
	}

	content := txt.ReplaceNoteText(existing, newText)
	if err := b.fs.Write(dir, filename, content); err != nil {
		return fmt.Errorf("edit note: %w", err)
	}

	return b.showFile([]string{dirHash, filenameHash})
}

func (b *Bot) showEditTask(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show edit task: missing params")
	}

	back := tg.NewCmd(CmdShowTask, params)
	kb := tg.NewKeyboard([]tg.Row{tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back))})

	cmd := tg.NewCmd(CmdEditTask, append(append([]string(nil), params...), "%s"))
	b.db.SetInputExpectation(cmd)

	return b.showHTML(i18n.Tr("OK. Send me the new text for your task"), kb)
}

func (b *Bot) editTask(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("edit task: missing params")
	}
	newText := strings.TrimSpace(params[len(params)-1])
	if newText == "" {
		return fmt.Errorf("edit task: empty text")
	}
	kind := params[0]

	switch kind {
	case taskKindChat:
		return b.editChatTask(params[1], newText)
	case taskKindList:
		if len(params) < 4 {
			return fmt.Errorf("edit task: missing checklist params")
		}
		return b.editListTask(params[1], params[2], newText)
	case taskKindArea:
		if len(params) < 4 {
			return fmt.Errorf("edit task: missing area params")
		}
		return b.editAreaTask(params[1], params[2], newText)
	default:
		return fmt.Errorf("edit task: unknown kind %q", kind)
	}
}

func (b *Bot) editChatTask(msgHash, newText string) error {
	chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("edit chat task: %w", err)
	}
	updated, newHash, err := editChatMsg(chatMD, msgHash, newText)
	if err != nil {
		return fmt.Errorf("edit chat task: %w", err)
	}
	if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, updated); err != nil {
		return fmt.Errorf("edit chat task: %w", err)
	}
	return b.showChatTask(newHash)
}

func (b *Bot) editListTask(checklistHash, itemHash, newText string) error {
	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("edit list task: %w", err)
	}
	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("edit list task: %w", err)
	}
	newMD, ok := txt.ReplaceChecklistItem(md, itemHash, newText)
	if !ok {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(fs.DirUserRoot, checklist, newMD); err != nil {
		return fmt.Errorf("edit list task: %w", err)
	}
	return b.showListTask(checklistHash, fs.Hash(newText))
}

func (b *Bot) editAreaTask(projectHash, itemHash, newText string) error {
	projectPath, err := b.fs.ResolveDirParam(projectHash)
	if err != nil {
		return fmt.Errorf("edit area task: %w", err)
	}
	md, err := b.fs.Read(projectPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("edit area task: %w", err)
	}
	newMD, ok := txt.ReplaceChecklistItem(md, itemHash, newText)
	if !ok {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(projectPath, life.TasksFilename, newMD); err != nil {
		return fmt.Errorf("edit area task: %w", err)
	}
	return b.showAreaTask(projectHash, fs.Hash(newText))
}

// editChatMsg replaces the body of a chat task block, preserving marker and timestamp.
func editChatMsg(content, msgHash, newBody string) (string, string, error) {
	blocks := readChatMsgs(content)
	idx := -1
	for i, block := range blocks {
		if inboxHeaderRegex.MatchString(block) {
			continue
		}
		if chatBlockHash(block) == msgHash {
			idx = i
			break
		}
	}
	if idx == -1 {
		return "", "", fmt.Errorf("chat block not found for hash %q", msgHash)
	}

	prefix := inboxEntryPrefix.FindString(blocks[idx])
	newBody = strings.TrimSpace(newBody)
	blocks[idx] = prefix + newBody

	return strings.Join(blocks, "\n"), chatBlockHash(blocks[idx]), nil
}

func editNoteCmd(dirHash, filename string) tg.Cmd {
	return tg.NewCmd(CmdShowEditNote, []string{dirHash, fs.Hash(filename)})
}

func editTaskCmd(params ...string) tg.Cmd {
	return tg.NewCmd(CmdShowEditTask, params)
}
