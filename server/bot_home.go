package server

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/morningsummary"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
)

func (b *Bot) moveToAreaTask(params []string) error {
	projectPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("move to area task: %w", err)
	}
	msgHash := params[1]

	err = b.moveFromChat(func(content string, _ time.Time) error {
		md, readErr := b.fs.Read(projectPath, life.TasksFilename)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return readErr
		}
		md = txt.AddChecklistItem(md, content, false)
		return b.fs.Write(projectPath, life.TasksFilename, md)
	}, true, msgHash)
	if err != nil {
		return fmt.Errorf("move to area task: %w", err)
	}

	b.delAllKeyboards()
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
	newMD, item := txt.RemoveChecklistItem(md, itemHash)
	if item == "" {
		return b.showTasksView(nil)
	}
	if err := b.fs.Write(projectPath, life.TasksFilename, newMD); err != nil {
		return fmt.Errorf("complete area task: %w", err)
	}
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
			areaKey = "inbox"
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
	msgHash string
	source  string
	areaKey string
	cmd     tg.Cmd
}

func (b *Bot) collectOpenTasks() []taskEntry {
	var entries []taskEntry

	chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		entries = append(entries, parseChatTasks(chatMD, fs.ChatFilename)...)
	}
	laterMD, err := b.fs.Read(fs.DirUserRoot, fs.LaterFilename)
	if err == nil {
		for _, t := range parseChecklistTasks(laterMD, fs.LaterFilename) {
			entries = append(entries, t)
		}
	}

	_ = life.EnsureSpheresRoot(b.fs)
	spheres, _ := life.ListSpheres(b.fs)
	for _, sphere := range spheres {
		projects, _ := life.ListProjects(b.fs, sphere)
		for _, project := range projects {
			tasksFile := fs.JoinDir(project, life.TasksFilename)
			md, err := b.fs.Read(project, life.TasksFilename)
			if err != nil {
				continue
			}
			for _, t := range parseChecklistTasks(md, tasksFile) {
				t.areaKey = project
				t.cmd = tg.NewCmd(CmdCompleteAreaTask, []string{fs.ShortHash(project), fs.Hash(t.text)})
				entries = append(entries, t)
			}
		}
	}

	return entries
}

var chatEntryRE = regexp.MustCompile("^- \\[ \\] (?:`\\d{2}:\\d{2}` )?(.*)$")

func parseChatTasks(md, source string) []taskEntry {
	var out []taskEntry
	for _, block := range readChatMsgs(md) {
		firstLine := strings.SplitN(block, "\n", 2)[0]
		m := chatEntryRE.FindStringSubmatch(firstLine)
		if m == nil {
			continue
		}
		hash := chatBlockHash(block)
		text := strings.TrimSpace(m[1])
		out = append(out, taskEntry{
			text:    text,
			msgHash: hash,
			source:  source,
			cmd:     tg.NewCmd(CmdComplete, []string{hash}),
		})
	}
	return out
}

func parseChecklistTasks(md, source string) []taskEntry {
	var out []taskEntry
	for _, item := range txt.IncompleteChecklistItems(md) {
		out = append(out, taskEntry{
			text:    item,
			msgHash: fs.Hash(item),
			source:  source,
			cmd:     tg.NewCmd(CmdCompleteChecklistItem, []string{fs.Hash(source), fs.Hash(item)}),
		})
	}
	return out
}

func taskBtn(t taskEntry) tg.Btn {
	label := t.text
	if len([]rune(label)) > maxHeaderLengthForMobile {
		label = txt.Substr(label, 0, maxHeaderLengthForMobile-1) + "…"
	}
	return tg.NewBtn(label, t.cmd)
}

func sortedAreaKeys(m map[string][]taskEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// inbox first, then alphabetical paths
	sort.Strings(keys)
	if idx := indexOf(keys, "inbox"); idx > 0 {
		keys = append([]string{"inbox"}, append(keys[:idx], keys[idx+1:]...)...)
	}
	return keys
}

func indexOf(ss []string, s string) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}

func areaLabelForKey(key string) string {
	if key == "inbox" {
		return "📥"
	}
	return life.AreaEmoji(key)
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
