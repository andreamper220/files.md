package server

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
)

const noPriorityEmoji = "⚪️"

const (
	draftKindTask = "t"
	draftKindNote = "n"
)

func (b *Bot) storePendingDraft(content string) string {
	hash := fs.Hash(fmt.Sprintf("%s:%d", content, time.Now().UnixNano()))
	b.db.SetPendingDraft(hash, content)
	return hash
}

func (b *Bot) takePendingDraft(hash string) (string, error) {
	content, ok := b.db.PendingDraft(hash)
	if !ok {
		return "", fmt.Errorf("pending draft %q not found", hash)
	}
	b.db.DelPendingDraft(hash)
	return content, nil
}

// showSaveType asks whether incoming content should become a task or a note.
func (b *Bot) showSaveType(params []string) error {
	draftHash := params[0]
	_, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn("📋", tg.NewCmd(CmdSaveAsTask, []string{draftHash})),
			tg.NewBtn("🗒", tg.NewCmd(CmdSaveAsNote, []string{draftHash})),
			tg.NewBtn("✕", tg.NewCmd(CmdCancelPendingDraft, []string{draftHash})),
		),
	})
	return b.showHTML(i18n.Tr("Choose: task or note?"), kb)
}

func (b *Bot) cancelPendingDraft(params []string) error {
	if len(params) > 0 {
		b.db.DelPendingDraft(params[0])
	}
	return b.ShowHome(nil)
}

func (b *Bot) saveAsTask(params []string) error {
	draftHash := params[0]
	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	if txt.IsVoiceDraft(content) {
		return b.showVoiceTitlePicker(draftHash, draftKindTask)
	}
	if txt.NeedsUserTitle(content) {
		return b.showDraftTitlePrompt(draftHash, draftKindTask)
	}
	return b.showTaskPriorityPicker(draftHash)
}

func (b *Bot) pickTaskPriority(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("pick task priority: missing params")
	}
	draftHash := params[0]
	if _, ok := b.db.PendingDraft(draftHash); !ok {
		return b.ShowHome(nil)
	}
	return b.showTaskAreaPicker(draftHash, params[1])
}

func (b *Bot) saveAsNote(params []string) error {
	draftHash := params[0]
	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	if txt.IsVoiceDraft(content) {
		return b.showVoiceTitlePicker(draftHash, draftKindNote)
	}
	if txt.NeedsUserTitle(content) {
		return b.showDraftTitlePrompt(draftHash, draftKindNote)
	}
	return b.showNoteAreaPicker(draftHash)
}

func (b *Bot) showTaskPriorityPicker(draftHash string) error {
	emojis := displayPriorityEmojis(b.cfg.PriorityEmojis())
	if len(emojis) == 0 {
		return b.showTaskAreaPicker(draftHash, "0")
	}

	kb := tg.NewKeyboard(nil)
	row := tg.NewRow()
	for i, emoji := range emojis {
		row = append(row, tg.NewBtn(emoji, tg.NewCmd(CmdPickTaskPriority, []string{draftHash, strconv.Itoa(i)})))
		if len(row) >= btnsPerRow {
			kb.AddRow(row)
			row = tg.NewRow()
		}
	}
	if len(row) > 0 {
		kb.AddRow(row)
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	return b.showHTML(i18n.Tr("Выбери срочность:"), kb)
}

func (b *Bot) showTaskAreaPicker(draftHash, priorityIdxStr string) error {
	_ = life.EnsureSpheresRoot(b.fs)

	kb := tg.NewKeyboard(nil)
	areas := b.collectAllAreas()
	addAreaPickerBtns(kb, areas, func(projectPath string) tg.Cmd {
		return tg.NewCmd(CmdSaveTaskToArea, []string{draftHash, fs.ShortHash(projectPath), priorityIdxStr})
	})
	if len(areas) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("🏗 Создать структуру"), tg.NewCmd(CmdInitLife, nil)))
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	return b.showHTML(i18n.Tr("Выбери область для задачи:"), kb)
}

func (b *Bot) showNoteAreaPicker(draftHash string) error {
	_ = life.EnsureSpheresRoot(b.fs)

	kb := tg.NewKeyboard(nil)
	areas := b.collectAllAreas()
	addAreaPickerBtns(kb, areas, func(projectPath string) tg.Cmd {
		return tg.NewCmd(CmdSaveNoteToArea, []string{draftHash, fs.ShortHash(projectPath), life.KindCode(life.KindDraft)})
	})
	if len(areas) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("🏗 Создать структуру"), tg.NewCmd(CmdInitLife, nil)))
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	return b.showHTML(i18n.Tr("Выбери область:"), kb)
}

func (b *Bot) saveTaskToArea(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("save task to area: missing priority")
	}
	draftHash := params[0]
	projectPath, err := b.fs.ResolveDirParam(params[1])
	if err != nil {
		return fmt.Errorf("save task to area: %w", err)
	}

	content, err := b.takePendingDraft(draftHash)
	if err != nil {
		return err
	}

	taskText, err := taskTextFromDraft(content)
	if err != nil {
		return fmt.Errorf("save task to area: %w", err)
	}

	if emoji, ok := priorityEmojiAt(b.cfg.PriorityEmojis(), params[2]); ok {
		taskText = priority.Apply(taskText, emoji, b.cfg.PriorityEmojis())
	}

	md, readErr := b.fs.Read(projectPath, life.TasksFilename)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("save task to area: %w", readErr)
	}
	md = txt.AddChecklistItem(md, taskText, false)
	if err := b.fs.Write(projectPath, life.TasksFilename, md); err != nil {
		return fmt.Errorf("save task to area: %w", err)
	}

	b.setRecentLifeProject(projectPath)
	b.delAllKeyboards()
	spherePath := life.SpherePathFromArea(projectPath)
	msg := fmt.Sprintf(i18n.Tr("Сохранено в <b>%s</b>"), life.SaveLocationLabel(spherePath, projectPath))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func displayPriorityEmojis(emojis []string) []string {
	out := make([]string, 0, len(emojis))
	for _, e := range emojis {
		if e == noPriorityEmoji {
			continue
		}
		out = append(out, e)
	}
	return out
}

func priorityEmojiAt(emojis []string, idxStr string) (string, bool) {
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return "", false
	}
	display := displayPriorityEmojis(emojis)
	if idx < 0 || idx >= len(display) {
		return "", false
	}
	return display[idx], true
}

func taskTextFromDraft(content string) (string, error) {
	if title := txt.DraftTitle(content); title != "" {
		return title, nil
	}
	content = strings.TrimSpace(txt.NormNewLines(content))
	if content == "" {
		return "", fmt.Errorf("empty task")
	}
	return "", fmt.Errorf("empty task")
}

func (b *Bot) showDraftTitlePrompt(draftHash, kind string) error {
	b.db.SetInputExpectation(tg.NewCmd(CmdApplyDraftTitle, []string{draftHash, kind, "%s"}))
	return b.showHTML(i18n.Tr("Пришли название:"), nil)
}

func (b *Bot) showVoiceTitlePicker(draftHash, kind string) error {
	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}

	suggestions := txt.VoiceTitleSuggestions(content)
	kb := tg.NewKeyboard(nil)
	row := tg.NewRow()
	for i, title := range suggestions {
		row = append(row, tg.NewBtn(
			txt.BtnLabelTitle(title, txt.VoiceTitleBtnMaxRunes),
			tg.NewCmd(CmdPickVoiceTitle, []string{draftHash, kind, strconv.Itoa(i)}),
		))
		if len(row) >= 2 {
			kb.AddRow(row)
			row = tg.NewRow()
		}
	}
	if len(row) > 0 {
		kb.AddRow(row)
	}
	kb.AddRow(tg.NewBtn("✏️ "+i18n.Tr("Своё"), tg.NewCmd(CmdVoiceTitleCustom, []string{draftHash, kind})))

	preview := txt.VoiceSummary(content)
	if preview == "" {
		preview = txt.VoicePlaceholder
	}
	msg := i18n.Tr("Выбери заголовок:") + "\n\n<i>" + txt.EscapeHTML(preview) + "</i>"
	return b.showHTML(msg, kb)
}

func (b *Bot) pickVoiceTitle(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("pick voice title: missing params")
	}
	draftHash := params[0]
	kind := params[1]
	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	idx, err := strconv.Atoi(params[2])
	if err != nil {
		return fmt.Errorf("pick voice title: bad index: %w", err)
	}
	suggestions := txt.VoiceTitleSuggestions(content)
	if idx < 0 || idx >= len(suggestions) {
		return b.ShowHome(nil)
	}
	return b.applyVoiceDraftTitle(draftHash, kind, suggestions[idx])
}

func (b *Bot) voiceTitleCustom(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("voice title custom: missing params")
	}
	return b.showDraftTitlePrompt(params[0], params[1])
}

func (b *Bot) applyVoiceDraftTitle(draftHash, kind, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return b.showVoiceTitlePicker(draftHash, kind)
	}
	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	b.db.SetPendingDraft(draftHash, txt.ApplyVoiceDraftTitle(content, title))
	b.db.DelInputExpectation()

	switch kind {
	case draftKindTask:
		return b.showTaskPriorityPicker(draftHash)
	case draftKindNote:
		return b.showNoteAreaPicker(draftHash)
	default:
		return b.ShowHome(nil)
	}
}

func (b *Bot) applyDraftTitle(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("apply draft title: missing params")
	}
	draftHash := params[0]
	kind := params[1]
	title := strings.TrimSpace(params[2])
	if title == "" {
		return b.showDraftTitlePrompt(draftHash, kind)
	}

	content, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	if txt.IsVoiceDraft(content) {
		b.db.SetPendingDraft(draftHash, txt.ApplyVoiceDraftTitle(content, title))
	} else {
		b.db.SetPendingDraft(draftHash, txt.ApplyDraftTitle(content, title))
	}
	b.db.DelInputExpectation()

	switch kind {
	case draftKindTask:
		return b.showTaskPriorityPicker(draftHash)
	case draftKindNote:
		return b.showNoteAreaPicker(draftHash)
	default:
		return b.ShowHome(nil)
	}
}

func (b *Bot) saveNoteToArea(params []string) error {
	draftHash := params[0]
	projectPath, err := b.fs.ResolveDirParam(params[1])
	if err != nil {
		return fmt.Errorf("save note to area: %w", err)
	}
	kind, ok := life.KindFromCode(params[2])
	if !ok {
		return fmt.Errorf("save note to area: bad kind")
	}

	content, err := b.takePendingDraft(draftHash)
	if err != nil {
		return err
	}

	docDir := life.DocDir(projectPath, kind)
	title, body, err := b.extractHeaderAndBodyPreserveMedia(content, maxHeaderLength)
	if err != nil {
		return fmt.Errorf("save note to area: %w", err)
	}

	filename := fs.Filename(title)
	if err := b.createOrAdd(docDir, filename, body); err != nil {
		return fmt.Errorf("save note to area: %w", err)
	}
	if err := life.RegisterDoc(b.fs, kind, docDir, filename); err != nil {
		return fmt.Errorf("save note to area: register: %w", err)
	}

	b.setRecentLifeProject(projectPath)
	b.delAllKeyboards()
	spherePath := life.SpherePathFromArea(projectPath)
	msg := fmt.Sprintf(i18n.Tr("Сохранено в <b>%s</b> → %s"), life.SaveLocationLabel(spherePath, projectPath), lifeKindLabel(kind))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) queueIncomingContent(content string) error {
	if b.cfg.JournalOnlyMode() {
		hash := b.storePendingDraft(content)
		return b.saveAsNote([]string{hash})
	}

	draftHash := b.storePendingDraft(content)
	return b.showSaveType([]string{draftHash})
}

func (b *Bot) showTaskActions(_ []string) error {
	return b.ShowHome(nil)
}
