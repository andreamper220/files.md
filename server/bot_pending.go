package server

import (
	"fmt"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/tg"
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
		),
	})
	return b.showHTML(i18n.Tr("Choose: task or note?"), kb)
}

func (b *Bot) saveAsTask(params []string) error {
	content, err := b.takePendingDraft(params[0])
	if err != nil {
		return err
	}

	msgHash, err := b.appendToChat(content, b.cfg.Timezone())
	if err != nil {
		return fmt.Errorf("save as task: %w", err)
	}

	return b.showTaskActions([]string{msgHash})
}

func (b *Bot) saveAsNote(params []string) error {
	draftHash := params[0]
	_, ok := b.db.PendingDraft(draftHash)
	if !ok {
		return b.ShowHome(nil)
	}
	return b.showNoteAreaPicker(draftHash)
}

func (b *Bot) showNoteAreaPicker(draftHash string) error {
	_ = life.EnsureSpheresRoot(b.fs)
	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return fmt.Errorf("note area picker: %w", err)
	}

	kb := tg.NewKeyboard(nil)
	for _, spherePath := range spheres {
		projects, err := life.ListProjects(b.fs, spherePath)
		if err != nil {
			continue
		}
		for _, projectPath := range projects {
			label := life.SphereEmoji(spherePath) + life.AreaEmoji(projectPath)
			kb.AddRow(tg.NewBtn(
				label,
				tg.NewCmd(CmdSaveNoteToArea, []string{draftHash, fs.ShortHash(projectPath), life.KindCode(life.KindDraft)}),
			))
		}
	}
	if len(spheres) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("🏗 Создать структуру"), tg.NewCmd(CmdInitLife, nil)))
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	return b.showHTML(i18n.Tr("Выбери область:"), kb)
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
	msg := fmt.Sprintf(i18n.Tr("Сохранено в <b>%s</b> → %s"), life.AreaEmoji(projectPath), lifeKindLabel(kind))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) queueIncomingContent(content string) error {
	if b.cfg.ChatOnlyMode() {
		_, err := b.appendToChat(content, b.cfg.Timezone())
		return err
	}
	if b.cfg.JournalOnlyMode() {
		hash := b.storePendingDraft(content)
		return b.saveAsNote([]string{hash})
	}

	draftHash := b.storePendingDraft(content)
	return b.showSaveType([]string{draftHash})
}

// showTaskActions shows compact emoji controls for a saved task.
func (b *Bot) showTaskActions(params []string) error {
	msgHash := params[0]

	kb := tg.NewKeyboard(nil)
	prioBtns := b.priorityBtns(msgHash)
	if len(prioBtns) > 0 {
		kb.AddRow(tg.NewRow(prioBtns...))
	}

	areaBtns := b.areaTaskBtns(msgHash)
	for _, row := range chunkBtns(areaBtns, btnsPerRow) {
		kb.AddRow(row)
	}

	kb.AddRow(tg.NewRow(
		tg.NewBtn("⏳", tg.NewCmd(CmdMoveToLater, []string{msgHash})),
		tg.NewBtn("🌚", tg.NewCmd(CmdScheduleForTmrw, []string{msgHash})),
		tg.NewBtn("🗒", tg.NewCmd(CmdMoveToDraft, []string{msgHash})),
		tg.NewBtn("🌐", tg.NewCmd(CmdShowLifeSpheres, nil)),
	))

	b.delAllKeyboards()
	return b.showHTML(i18n.Tr("Saved!"), kb)
}

func (b *Bot) areaTaskBtns(msgHash string) []tg.Btn {
	_ = life.EnsureSpheresRoot(b.fs)
	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return nil
	}

	var btns []tg.Btn
	for _, spherePath := range spheres {
		projects, err := life.ListProjects(b.fs, spherePath)
		if err != nil {
			continue
		}
		for _, projectPath := range projects {
			label := life.SphereEmoji(spherePath) + life.AreaEmoji(projectPath)
			btns = append(btns, tg.NewBtn(
				label,
				tg.NewCmd(CmdMoveToAreaTask, []string{fs.ShortHash(projectPath), msgHash}),
			))
		}
	}
	return btns
}

func chunkBtns(btns []tg.Btn, size int) [][]tg.Btn {
	if len(btns) == 0 {
		return nil
	}
	var rows [][]tg.Btn
	for i := 0; i < len(btns); i += size {
		end := i + size
		if end > len(btns) {
			end = len(btns)
		}
		rows = append(rows, btns[i:end])
	}
	return rows
}
