package server

import (
	"errors"
	"fmt"
	"os"
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

	kb := taskDetailKeyboard(taskDetailOpts{
		back:     tg.NewCmd(CmdShowTasksView, nil),
		complete: tg.NewCmd(CmdComplete, []string{msgHash}),
		delete:   tg.NewCmd(CmdDeleteTask, []string{taskKindChat, msgHash}),
	})
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

	kb := taskDetailKeyboard(taskDetailOpts{
		back:     back,
		complete: tg.NewCmd(CmdCompleteChecklistItem, []string{checklistHash, itemHash}),
		delete:   tg.NewCmd(CmdDeleteTask, []string{taskKindList, checklistHash, itemHash}),
	})
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

	moveCmd := tg.NewCmd(CmdMoveToAreaTask, []string{projectHash, itemHash})
	kb := taskDetailKeyboard(taskDetailOpts{
		back:        tg.NewCmd(CmdShowTasksView, nil),
		complete:    tg.NewCmd(CmdCompleteAreaTask, []string{projectHash, itemHash}),
		delete:      tg.NewCmd(CmdDeleteTask, []string{taskKindArea, projectHash, itemHash}),
		move:        &moveCmd,
		projectPath: projectPath,
	})
	title := fmt.Sprintf("%s %s", life.AreaEmoji(projectPath), life.AreaFullTitle(projectPath))
	displayItem := txt.VoiceDetailBody(item)
	if displayItem == "" {
		displayItem = txt.DisplayText(item)
	}
	if displayItem == "" {
		displayItem = item
	}
	return b.showMD(title+"\n\n"+displayItem, kb)
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

type taskDetailOpts struct {
	back        tg.Cmd
	complete    tg.Cmd
	delete      tg.Cmd
	move        *tg.Cmd
	projectPath string
}

func taskDetailKeyboard(opts taskDetailOpts) *tg.Keyboard {
	row := tg.NewRow(
		tg.NewBtn("⬅️", opts.back),
		tg.NewBtn("🔎", tg.NewCustomCmd(CmdInlineQuerySearchEveryWhere, nil, tg.CmdTypeInlineQueryCurrentChat)),
		tg.NewBtn("✅", opts.complete),
	)
	if opts.move != nil {
		row = append(row, tg.NewBtn("↔️", *opts.move))
	}
	row = append(row, tg.NewBtn("🗑", opts.delete))
	if opts.projectPath != "" {
		row = append(row, tg.NewBtn(
			life.AreaNavBtnLabel(opts.projectPath),
			tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(opts.projectPath)}),
		))
	}
	row = append(row, tg.NewBtn("🏠", tg.NewCmd(CmdShowHome, nil)))
	return tg.NewKeyboard([]tg.Row{row})
}

func (b *Bot) showMoveAreaTask(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show move area task: missing params")
	}
	srcHash := params[0]
	itemHash := params[1]

	_ = life.EnsureSpheresRoot(b.fs)

	var kb tg.Keyboard
	var dstAreas []string
	for _, areaPath := range b.collectAllAreas() {
		if fs.ShortHash(areaPath) == srcHash {
			continue
		}
		dstAreas = append(dstAreas, areaPath)
	}
	addAreaPickerBtns(&kb, dstAreas, func(areaPath string) tg.Cmd {
		return tg.NewCmd(CmdMoveAreaTask, []string{srcHash, itemHash, fs.ShortHash(areaPath)})
	})
	if len(dstAreas) == 0 {
		kb.AddRow(tg.NewBtn("—", tg.NewCmd(CmdDoNothing, nil)))
	}
	kb.AddRow(tg.NewBtn("⬅️", tg.NewCmd(CmdShowTask, []string{taskKindArea, srcHash, itemHash})))

	return b.showHTML(i18n.Tr("Переместить задачу в область:"), &kb)
}

func (b *Bot) moveAreaTask(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("move area task: missing params")
	}
	srcPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("move area task: %w", err)
	}
	dstPath, err := b.fs.ResolveDirParam(params[2])
	if err != nil {
		return fmt.Errorf("move area task: %w", err)
	}
	itemHash := params[1]

	srcMD, err := b.fs.Read(srcPath, life.TasksFilename)
	if err != nil {
		return fmt.Errorf("move area task: %w", err)
	}
	newSrcMD, item := txt.RemoveChecklistItem(srcMD, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}

	dstMD, readErr := b.fs.Read(dstPath, life.TasksFilename)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("move area task: %w", readErr)
	}
	dstMD = txt.AddChecklistItem(dstMD, item, false)

	if err := b.fs.Write(srcPath, life.TasksFilename, newSrcMD); err != nil {
		return fmt.Errorf("move area task: %w", err)
	}
	if err := b.fs.Write(dstPath, life.TasksFilename, dstMD); err != nil {
		return fmt.Errorf("move area task: %w", err)
	}

	return b.showTasksView(nil)
}

func taskPreviewLabel(text string) string {
	label := strings.TrimSpace(txt.DisplayText(text))
	if label == "" {
		label = strings.TrimSpace(text)
	}
	if label == "" {
		return "…"
	}
	if len([]rune(label)) > maxHeaderLengthForMobile {
		return txt.Substr(label, 0, maxHeaderLengthForMobile-1) + "…"
	}
	return label
}

func (b *Bot) moveToAreaTask(params []string) error {
	return b.showMoveAreaTask(params)
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
	return b.showHTML(tasksViewMessage(byArea), kbPtr)
}

func tasksViewMessage(byArea map[string][]taskEntry) string {
	lines := []string{i18n.Tr("📋 Tasks")}
	for _, key := range sortedAreaKeys(byArea) {
		tasks := byArea[key]
		if len(tasks) == 0 {
			continue
		}
		lines = append(lines, life.AreaListPrefix+areaLabelForKey(key))
	}
	return strings.Join(lines, "\n")
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
		areas, _ := life.ListAllAreas(b.fs, sphere)
		for _, project := range areas {
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
	return life.AreaEmoji(key) + " " + life.AreaFullTitle(key)
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
		areas, _ := life.ListAllAreas(b.fs, spherePath)
		addAreaNavBtns(kb, areas, CmdShowLifeProject)
	}

	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))
	return b.showHTML(report, kb)
}
