package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/journal"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/morningsummary"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
)

const (
	taskKindChat = "c"
	taskKindList = "k"
	taskKindArea = "a"
)

func (b *Bot) showTask(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show task: missing params")
	}
	switch params[0] {
	case taskKindChat:
		return b.showChatTask(params[1])
	case taskKindList:
		if len(params) < 3 {
			return fmt.Errorf("show task: missing checklist params")
		}
		return b.showListTask(params[1], params[2])
	case taskKindArea:
		if len(params) < 3 {
			return fmt.Errorf("show task: missing area params")
		}
		return b.showAreaTask(params[1], params[2])
	default:
		return fmt.Errorf("show task: unknown kind %q", params[0])
	}
}

func (b *Bot) deleteTask(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("delete task: missing params")
	}
	switch params[0] {
	case taskKindChat:
		return b.deleteChatTask(params[1])
	case taskKindList:
		if len(params) < 3 {
			return fmt.Errorf("delete task: missing checklist params")
		}
		return b.deleteListTask(params[1], params[2])
	case taskKindArea:
		if len(params) < 3 {
			return fmt.Errorf("delete task: missing area params")
		}
		return b.deleteAreaTask(params[1], params[2])
	default:
		return fmt.Errorf("delete task: unknown kind %q", params[0])
	}
}

func (b *Bot) showChatTask(msgHash string) error {
	chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("show chat task: %w", err)
	}
	_, block, ok := findChatMsgByHash(chatMD, msgHash)
	if !ok {
		return b.showTasksView(nil)
	}

	kb := taskDetailKeyboard(
		tg.NewCmd(CmdShowTasksView, nil),
		tg.NewCmd(CmdComplete, []string{msgHash}),
		tg.NewCmd(CmdDeleteTask, []string{taskKindChat, msgHash}),
	)
	if err := b.showMD(chatTaskDisplayBody(block), kb); err != nil {
		return fmt.Errorf("show chat task: %w", err)
	}

	msgID, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if hasLastKeyboard {
		b.db.SetHashOrPathByMsgID(msgID, "#"+msgHash)
	}
	return nil
}

func (b *Bot) showListTask(checklistHash, itemHash string) error {
	return b.showListTaskWithBack(checklistHash, itemHash, tg.NewCmd(CmdShowTasksView, nil))
}

func (b *Bot) showListTaskWithBack(checklistHash, itemHash string, back tg.Cmd) error {
	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("show list task: %w", err)
	}
	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("show list task: %w", err)
	}
	item := txt.ChecklistItem(md, itemHash)
	if item == "" {
		return b.dispatchBack(back)
	}

	kb := taskDetailKeyboard(
		back,
		tg.NewCmd(CmdCompleteChecklistItem, []string{checklistHash, itemHash}),
		tg.NewCmd(CmdDeleteTask, []string{taskKindList, checklistHash, itemHash}),
	)
	return b.showMD(item, kb)
}

func (b *Bot) dispatchBack(back tg.Cmd) error {
	switch back.Name {
	case CmdShowTasksView:
		return b.showTasksView(nil)
	case CmdShowChecklist:
		return b.showChecklist(back.Params)
	default:
		return b.ShowHome(nil)
	}
}

func (b *Bot) showAreaTask(projectHash, itemHash string) error {
	projectPath, err := b.fs.ResolveDirParam(projectHash)
	if err != nil {
		return fmt.Errorf("show area task: %w", err)
	}
	md, err := b.fs.Read(projectPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("show area task: %w", err)
	}
	item := txt.ChecklistItem(md, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}

	kb := taskDetailKeyboard(
		tg.NewCmd(CmdShowTasksView, nil),
		tg.NewCmd(CmdCompleteAreaTask, []string{projectHash, itemHash}),
		tg.NewCmd(CmdDeleteTask, []string{taskKindArea, projectHash, itemHash}),
	)
	return b.showMD(item, kb)
}

func (b *Bot) deleteChatTask(msgHash string) error {
	key, err := b.fs.SafePath(fs.DirUserRoot, "")
	if err != nil {
		return fmt.Errorf("delete chat task: %w", err)
	}
	lock := userLock(key)
	lock.Lock()
	defer lock.Unlock()

	content, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("delete chat task: %w", err)
	}
	newContent, removed, err := deleteChatMsg(content, msgHash)
	if err != nil {
		return fmt.Errorf("delete chat task: %w", err)
	}
	if !removed {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, newContent); err != nil {
		return fmt.Errorf("delete chat task: %w", err)
	}
	return b.showTasksView(nil)
}

func (b *Bot) deleteListTask(checklistHash, itemHash string) error {
	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("delete list task: %w", err)
	}
	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("delete list task: %w", err)
	}
	newMD, item := txt.RemoveChecklistItem(md, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(fs.DirUserRoot, checklist, newMD); err != nil {
		return fmt.Errorf("delete list task: %w", err)
	}
	return b.showTasksView(nil)
}

func (b *Bot) deleteAreaTask(projectHash, itemHash string) error {
	projectPath, err := b.fs.ResolveDirParam(projectHash)
	if err != nil {
		return fmt.Errorf("delete area task: %w", err)
	}
	md, err := b.fs.Read(projectPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("delete area task: %w", err)
	}
	newMD, item := txt.RemoveChecklistItem(md, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(projectPath, life.TasksFilename, newMD); err != nil {
		return fmt.Errorf("delete area task: %w", err)
	}
	return b.showTasksView(nil)
}

func taskDetailKeyboard(back, complete, delete tg.Cmd) *tg.Keyboard {
	return tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrBack), back),
			tg.NewBtn(i18n.Tr(i18n.StrComplete), complete),
			tg.NewBtn(i18n.Tr(i18n.StrDelete), delete),
		),
	})
}

func taskPreviewLabel(text string) string {
	label := strings.TrimSpace(text)
	if label == "" {
		return "…"
	}
	if len([]rune(label)) > maxHeaderLengthForMobile {
		return txt.Substr(label, 0, maxHeaderLengthForMobile-1) + "…"
	}
	return label
}

func (b *Bot) moveToAreaTask(params []string) error {
	// Legacy: tasks no longer pass through Chat.md; new tasks use CmdSaveTaskToArea.
	return b.ShowHome(nil)
}

func (b *Bot) completeAreaTask(params []string) error {
	projectPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("complete area task: %w", err)
	}
	itemHash := params[1]

	md, err := b.fs.Read(projectPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("complete area task: %w", err)
	}
	newMD, item := txt.CompleteChecklistItem(md, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(projectPath, life.TasksFilename, newMD); err != nil {
		return fmt.Errorf("complete area task: %w", err)
	}

	_ = journal.AddRecord(b.fs, fmt.Sprintf("✅ %s", fs.DisplayName(item)), b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())

	return b.showTasksView(nil)
}

func (b *Bot) showTasksView(_ []string) error {
	emojis := b.cfg.PriorityEmojis()
	byPriority := map[string][]taskEntry{}
	byArea := map[string][]taskEntry{}

	for _, entry := range b.collectOpenTasks() {
		p := priority.Detect(entry.text, emojis)
		if p == "" {
			p = "⚪️"
		}
		byPriority[p] = append(byPriority[p], entry)

		areaKey := entry.areaKey
		if areaKey == "" {
			continue
		}
		byArea[areaKey] = append(byArea[areaKey], entry)
	}

	var kb tg.Keyboard
	kbPtr := &kb
	prioOrder := append([]string{}, emojis...)
	prioOrder = append(prioOrder, "⚪️")

	for _, emoji := range prioOrder {
		tasks := byPriority[emoji]
		if len(tasks) == 0 {
			continue
		}
		row := tg.NewRow(tg.NewBtn(emoji, tg.NewCmd(CmdDoNothing, nil)))
		for _, t := range tasks {
			row = append(row, taskBtn(t))
			if len(row) >= btnsPerRow {
				kbPtr.AddRow(row)
				row = tg.NewRow(tg.NewBtn(emoji, tg.NewCmd(CmdDoNothing, nil)))
			}
		}
		if len(row) > 1 {
			kbPtr.AddRow(row)
		}
	}

	kbPtr.AddRow(tg.NewBtn("—", tg.NewCmd(CmdDoNothing, nil)))

	areaKeys := sortedAreaKeys(byArea)
	for _, key := range areaKeys {
		tasks := byArea[key]
		if len(tasks) == 0 {
			continue
		}
		row := tg.NewRow(tg.NewBtn(areaLabelForKey(key), tg.NewCmd(CmdDoNothing, nil)))
		for _, t := range tasks {
			row = append(row, taskBtn(t))
			if len(row) >= btnsPerRow {
				kbPtr.AddRow(row)
				row = tg.NewRow(tg.NewBtn(areaLabelForKey(key), tg.NewCmd(CmdDoNothing, nil)))
			}
		}
		if len(row) > 1 {
			kbPtr.AddRow(row)
		}
	}

	kbPtr.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))
	return b.showHTML(i18n.Tr("📋 Tasks"), kbPtr)
}

type taskEntry struct {
	text    string
	areaKey string
	cmd     tg.Cmd
}

func (b *Bot) collectOpenTasks() []taskEntry {
	var entries []taskEntry

	_ = life.EnsureSpheresRoot(b.fs)
	spheres, _ := life.ListSpheres(b.fs)
	for _, sphere := range spheres {
		projects, _ := life.ListProjects(b.fs, sphere)
		for _, project := range projects {
			md, err := b.fs.Read(project, life.TasksFilename)
			if err != nil {
				continue
			}
			for _, t := range parseAreaTasks(md, project) {
				entries = append(entries, t)
			}
		}
	}

	return entries
}

func parseAreaTasks(md, projectPath string) []taskEntry {
	var out []taskEntry
	for _, item := range txt.IncompleteChecklistItems(md) {
		out = append(out, taskEntry{
			text:    item,
			areaKey: projectPath,
			cmd:     tg.NewCmd(CmdShowTask, []string{taskKindArea, fs.ShortHash(projectPath), fs.Hash(item)}),
		})
	}
	return out
}

func taskBtn(t taskEntry) tg.Btn {
	return tg.NewBtn(taskPreviewLabel(t.text), t.cmd)
}

func sortedAreaKeys(m map[string][]taskEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func areaLabelForKey(key string) string {
	return life.AreaEmoji(key) + " " + life.AreaTitle(key)
}

func (b *Bot) showNotesHub(_ []string) error {
	report, err := morningsummary.BuildNotesHub(b.fs, b.cfg)
	if err != nil {
		return fmt.Errorf("notes hub: %w", err)
	}

	kb := tg.NewKeyboard(nil)
	_ = life.EnsureSpheresRoot(b.fs)
	spheres, _ := life.ListSpheres(b.fs)
	for _, spherePath := range spheres {
		projects, _ := life.ListProjects(b.fs, spherePath)
		row := tg.NewRow()
		for _, projectPath := range projects {
			label := life.SphereEmoji(spherePath) + life.AreaEmoji(projectPath)
			row = append(row, tg.NewBtn(label, tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(projectPath)})))
			if len(row) >= btnsPerRow {
				kb.AddRow(row)
				row = tg.NewRow()
			}
		}
		if len(row) > 0 {
			kb.AddRow(row)
		}
	}

	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))
	return b.showHTML(report, kb)
}
