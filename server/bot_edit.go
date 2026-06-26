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

const (
	editModeReplace = "r"
	editModeAppend  = "a"
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
	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr("✏️ Rewrite"), tg.NewCmd(CmdEditNoteReplace, params)),
			tg.NewBtn(i18n.Tr("➕ Append"), tg.NewCmd(CmdEditNoteAppend, params)),
		),
		tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back)),
	})

	return b.showHTML(i18n.Tr("How do you want to edit the note?"), kb)
}

func (b *Bot) startEditNoteReplace(params []string) error {
	return b.startEditNote(editModeReplace, params)
}

func (b *Bot) startEditNoteAppend(params []string) error {
	return b.startEditNote(editModeAppend, params)
}

func (b *Bot) startEditNote(mode string, params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("start edit note: missing params")
	}
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("start edit note: %w", err)
	}
	if _, err := b.fs.Unhash(dir, filenameHash); err != nil {
		return fmt.Errorf("start edit note: %w", err)
	}

	b.db.SetEditNoteTarget(dirHash, filenameHash, mode)

	back := tg.NewCmd(CmdCancelEditNote, []string{dirHash, filenameHash})
	kb := tg.NewKeyboard([]tg.Row{tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back))})

	msg := i18n.Tr("OK. Send new text or files to replace your note")
	if mode == editModeAppend {
		msg = i18n.Tr("OK. Send new text or attach files to your note")
	}
	return b.showHTML(msg, kb)
}

func (b *Bot) cancelEditNote(params []string) error {
	b.db.DelEditNoteTarget()
	if len(params) < 2 {
		return b.ShowHome(nil)
	}
	return b.showFile(params)
}

// editNote handles legacy input-expectation text edits.
func (b *Bot) editNote(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("edit note: missing params")
	}
	return b.editNoteText(params[0], params[1], params[2])
}

func (b *Bot) editNoteText(dirHash, filenameHash, newText string) error {
	return b.replaceEditNote(dirHash, filenameHash, newText)
}

func (b *Bot) replaceEditNote(dirHash, filenameHash, content string) error {
	content = txt.ReplaceNote(content)
	if content == "" {
		return fmt.Errorf("edit note: empty note")
	}

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("edit note: %w", err)
	}
	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("edit note: %w", err)
	}

	if err := b.fs.Write(dir, filename, content); err != nil {
		return fmt.Errorf("edit note: %w", err)
	}

	b.db.DelEditNoteTarget()
	return b.showFile([]string{dirHash, filenameHash})
}

func (b *Bot) applyEditNoteContent(dirHash, filenameHash, mode, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if mode == editModeAppend {
		return b.appendToEditNote(dirHash, filenameHash, content)
	}
	return b.replaceEditNote(dirHash, filenameHash, content)
}

func (b *Bot) appendToEditNote(dirHash, filenameHash, addition string) error {
	addition = strings.TrimSpace(addition)
	if addition == "" {
		return nil
	}

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("append to edit note: %w", err)
	}
	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("append to edit note: %w", err)
	}

	existing, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("append to edit note: %w", err)
	}

	content := txt.AppendNoteContent(existing, addition)
	if err := b.fs.Write(dir, filename, content); err != nil {
		return fmt.Errorf("append to edit note: %w", err)
	}

	return b.showFile([]string{dirHash, filenameHash})
}

func (b *Bot) showEditTask(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show edit task: missing params")
	}

	back := tg.NewCmd(CmdShowTask, params)
	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr("✏️ Rewrite"), tg.NewCmd(CmdEditTaskReplace, params)),
			tg.NewBtn(i18n.Tr("➕ Append"), tg.NewCmd(CmdEditTaskAppend, params)),
		),
		tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back)),
	})

	return b.showHTML(i18n.Tr("How do you want to edit the task?"), kb)
}

func (b *Bot) startEditTaskReplace(params []string) error {
	return b.startEditTask(editModeReplace, params)
}

func (b *Bot) startEditTaskAppend(params []string) error {
	return b.startEditTask(editModeAppend, params)
}

func (b *Bot) startEditTask(mode string, params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("start edit task: missing params")
	}

	b.db.SetEditTaskTarget(params, mode)

	back := tg.NewCmd(CmdShowTask, params)
	kb := tg.NewKeyboard([]tg.Row{tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), back))})

	msg := i18n.Tr("OK. Send new text or files to replace your task")
	if mode == editModeAppend {
		msg = i18n.Tr("OK. Send new text or attach files to your task")
	}
	return b.showHTML(msg, kb)
}

func (b *Bot) applyEditTaskContent(params []string, mode, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if err := b.editTask(append(append(append([]string(nil), params...), mode), content)); err != nil {
		return err
	}
	if mode == editModeReplace {
		b.db.DelEditTaskTarget()
	}
	return nil
}

func (b *Bot) editTask(params []string) error {
	if len(params) < 4 {
		return fmt.Errorf("edit task: missing params")
	}
	newText := strings.TrimSpace(params[len(params)-1])
	if newText == "" {
		return fmt.Errorf("edit task: empty text")
	}
	mode := params[len(params)-2]
	taskParams := params[:len(params)-2]
	kind := taskParams[0]

	switch kind {
	case taskKindChat:
		if mode == editModeAppend {
			return b.appendChatTask(taskParams[1], newText)
		}
		return b.editChatTask(taskParams[1], newText)
	case taskKindList:
		if len(taskParams) < 3 {
			return fmt.Errorf("edit task: missing checklist params")
		}
		if mode == editModeAppend {
			return b.appendListTask(taskParams[1], taskParams[2], newText)
		}
		return b.editListTask(taskParams[1], taskParams[2], newText)
	case taskKindArea:
		if len(taskParams) < 3 {
			return fmt.Errorf("edit task: missing area params")
		}
		if mode == editModeAppend {
			return b.appendAreaTask(taskParams[1], taskParams[2], newText)
		}
		return b.editAreaTask(taskParams[1], taskParams[2], newText)
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

func (b *Bot) appendChatTask(msgHash, addition string) error {
	chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("append chat task: %w", err)
	}
	updated, err := appendToChatMsg(chatMD, msgHash, addition)
	if err != nil {
		return fmt.Errorf("append chat task: %w", err)
	}
	if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, updated); err != nil {
		return fmt.Errorf("append chat task: %w", err)
	}
	return b.showChatTask(msgHash)
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
	return b.showListTask(checklistHash, checklistItemHash(newText))
}

func (b *Bot) appendListTask(checklistHash, itemHash, addition string) error {
	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("append list task: %w", err)
	}
	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("append list task: %w", err)
	}
	newMD, ok := txt.AppendChecklistItem(md, itemHash, addition)
	if !ok {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(fs.DirUserRoot, checklist, newMD); err != nil {
		return fmt.Errorf("append list task: %w", err)
	}
	return b.showListTask(checklistHash, itemHash)
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
	return b.showAreaTask(projectHash, checklistItemHash(newText))
}

func (b *Bot) appendAreaTask(projectHash, itemHash, addition string) error {
	projectPath, err := b.fs.ResolveDirParam(projectHash)
	if err != nil {
		return fmt.Errorf("append area task: %w", err)
	}
	md, err := b.fs.Read(projectPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("append area task: %w", err)
	}
	newMD, ok := txt.AppendChecklistItem(md, itemHash, addition)
	if !ok {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(projectPath, life.TasksFilename, newMD); err != nil {
		return fmt.Errorf("append area task: %w", err)
	}
	return b.showAreaTask(projectHash, itemHash)
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

func checklistItemHash(text string) string {
	first := strings.TrimSpace(strings.SplitN(txt.NormNewLines(text), "\n", 2)[0])
	return fs.Hash(first)
}
