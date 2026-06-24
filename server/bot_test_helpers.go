package server

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/db"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/morningsummary"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
	"github.com/zakirullin/files.md/server/userconfig"
)

func newTestBot(tgram *tg.FakeTG, userFS *fs.FS, cfg *userconfig.Config) (*Bot, *db.FakeDB) {
	fakeDB := db.NewFakeDB()
	return NewBot(-1, tgram, userFS, fakeDB, cfg), fakeDB
}

func pendingDraftHash(fakeDB *db.FakeDB) string {
	for h := range fakeDB.PendingDrafts {
		return h
	}
	return ""
}

func saveIncomingAsTask(bot *Bot, fakeDB *db.FakeDB) error {
	projectPath, err := ensureTestProject(bot)
	if err != nil {
		return err
	}
	h := pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("no pending draft")
	}
	if err := bot.saveAsTask([]string{h}); err != nil {
		return err
	}
	h = pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("pending draft gone after saveAsTask")
	}
	if err := ensureDraftTitled(bot, h, draftKindTask); err != nil {
		return err
	}
	if err := bot.pickTaskPriority([]string{h, "0"}); err != nil {
		return err
	}
	h = pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("pending draft gone after pickTaskPriority")
	}
	return bot.saveTaskToArea([]string{h, fs.ShortHash(projectPath), "0"})
}

// saveIncomingToInbox keeps legacy Chat.md flows working in tests (move/schedule panels).
func saveIncomingToInbox(bot *Bot, fakeDB *db.FakeDB) error {
	h := pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("no pending draft")
	}
	content, err := bot.takePendingDraft(h)
	if err != nil {
		return err
	}
	_, err = bot.appendToChat(content, bot.cfg.Timezone())
	return err
}

func ensureTestProject(bot *Bot) (string, error) {
	if err := life.Init(bot.fs); err != nil {
		return "", err
	}
	spherePath := life.SpherePath("Личное")
	projects, err := life.ListProjects(bot.fs, spherePath)
	if err != nil {
		return "", err
	}
	if len(projects) > 0 {
		return projects[0], nil
	}
	return life.CreateProject(bot.fs, spherePath, "test")
}

func tasksMDForTestProject(bot *Bot) (string, error) {
	projectPath, err := ensureTestProject(bot)
	if err != nil {
		return "", err
	}
	return bot.fs.Read(projectPath, life.TasksFilename)
}

func saveIncomingAsNoteToProject(bot *Bot, fakeDB *db.FakeDB, projectPath string) error {
	h := pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("no pending draft")
	}
	if err := bot.saveAsNote([]string{h}); err != nil {
		return err
	}
	h = pendingDraftHash(fakeDB)
	if h == "" {
		return fmt.Errorf("pending draft gone after saveAsNote")
	}
	if err := ensureDraftTitled(bot, h, draftKindNote); err != nil {
		return err
	}
	return bot.saveNoteToArea([]string{h, fs.ShortHash(projectPath), life.KindCode(life.KindDraft)})
}

func ensureDraftTitled(bot *Bot, draftHash, kind string) error {
	content, ok := bot.db.PendingDraft(draftHash)
	if !ok || !txt.NeedsUserTitle(content) {
		return nil
	}
	return bot.applyDraftTitle([]string{draftHash, kind, defaultTestDraftTitle(content, kind)})
}

func defaultTestDraftTitle(content, kind string) string {
	if title := txt.DraftTitle(content); title != "" {
		return title
	}
	if txt.HasImage(content) {
		if kind == draftKindNote {
			return fmt.Sprintf(i18n.Tr("Img %s"), now().Format("02.01.06 15:04"))
		}
		return "Photo"
	}
	if att, ok := txt.ParseAttachment(content); ok {
		return txt.AttachmentDisplayName(att.Name, att.Path)
	}
	return "Draft"
}

func initLifeTestProject(t *testing.T, userFS *fs.FS) string {
	t.Helper()
	r := require.New(t)
	r.NoError(life.Init(userFS))
	projectPath, err := life.CreateProject(userFS, life.SpherePath("Личное"), "test")
	r.NoError(err)
	return projectPath
}

func lifeDraftsDir(projectPath string) string {
	return life.DocDir(projectPath, life.KindDraft)
}

func homeNavKeyboard() *tg.Keyboard {
	return tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr("📋 Tasks"), tg.NewCmd(CmdShowTasksView, nil)),
			tg.NewBtn(i18n.Tr("🗒 Notes"), tg.NewCmd(CmdShowNotesHub, nil)),
			tg.NewBtn(i18n.Tr("🌐 Spheres"), tg.NewCmd(CmdShowLifeSpheres, nil)),
		),
	})
}

func emptyHomeText() string {
	return "🌴 " + i18n.Tr("Nothing here yet - send me something!")
}

func homeMessageText(userFS *fs.FS, cfg *userconfig.Config, shownCount int) string {
	report, err := morningsummary.Build(userFS, cfg)
	if err != nil {
		return "🌴"
	}
	report = strings.TrimSpace(report)
	if report == "" {
		return "🌴"
	}
	return report
}

func homeCountText(n int) string {
	postfix := i18n.Tr("items")
	if n == 1 {
		postfix = i18n.Tr("item")
	}
	return fmt.Sprintf(i18n.Tr("<b>%d</b> %s%s"), n, postfix, wideSpacer)
}

func homeButton() tg.Btn {
	return tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))
}

func allKeyboardBtnNames(kb *tg.Keyboard) []string {
	if kb == nil {
		return nil
	}
	var names []string
	for _, row := range kb.Btns {
		switch v := row.(type) {
		case tg.Btn:
			names = append(names, v.Name)
		case []tg.Btn:
			for _, btn := range v {
				names = append(names, btn.Name)
			}
		}
	}
	return names
}

func keyboardHasTaskNamed(kb *tg.Keyboard, name string) bool {
	for _, n := range allKeyboardBtnNames(kb) {
		if strings.Contains(n, name) {
			return true
		}
	}
	return false
}

func addQuickBtn(label, cmd string) tg.Btn {
	return tg.NewBtn(
		fmt.Sprintf("%s %s ➕", i18n.Emoji(label), i18n.Tr(label)),
		tg.NewCmd(CmdAddToQuickBtns, []string{cmd}),
	)
}

func delQuickBtn(label, cmd string) tg.Btn {
	return tg.NewBtn(
		fmt.Sprintf("%s %s ➖", i18n.Emoji(label), i18n.Tr(label)),
		tg.NewCmd(CmdDelFromQuickBtns, []string{cmd}),
	)
}

func configureQuickButtonsText() string {
	return fmt.Sprintf(
		i18n.Tr("Configure quick buttons (%s = add to quick buttons, %s = to remove from quick buttons):"),
		"➕", "➖",
	)
}

func fullSaveKeyboard(h string) *tg.Keyboard {
	var kb tg.Keyboard

	prioRow := tg.NewRow()
	for i, emoji := range priority.DefaultEmojis {
		prioRow = append(prioRow, tg.NewBtn(emoji, tg.NewCmd(CmdSetPriority, []string{h, strconv.Itoa(i)})))
	}
	kb.AddRow(prioRow)

	var catBtns []tg.Btn
	for _, category := range userconfig.DefaultConfig.TaskCategories {
		filename := fs.SanitizeFilename(category) + "_.md"
		catBtns = append(catBtns, tg.NewBtn(category, tg.NewCmd(CmdMoveToChecklist, []string{fs.Hash(filename), h})))
	}
	kb.AddRow(tg.NewRow(catBtns[0:3]...))
	kb.AddRow(tg.NewRow(catBtns[3:5]...))

	kb.AddRow(tg.NewRow(
		tg.NewBtn(i18n.Tr(i18n.StrToTomorrow), tg.NewCmd(CmdScheduleForTmrw, []string{h})),
		tg.NewBtn(i18n.Tr(i18n.StrToLater), tg.NewCmd(CmdMoveToLater, []string{h})),
		tg.NewBtn(i18n.Tr(i18n.StrToADay), tg.NewCmd(CmdShowScheduleForDay, []string{h})),
	))
	kb.AddRow(tg.NewRow(
		tg.NewBtn(i18n.Tr(i18n.StrToFile), tg.NewCmd(CmdShowMoveToDirOrFile, []string{h})),
		tg.NewBtn(i18n.Tr(i18n.StrToJournal), tg.NewCmd(CmdMoveToJournal, []string{h})),
		tg.NewBtn("👌", tg.NewCmd(CmdShowHome, []string{})),
	))

	return &kb
}

// tasksContainsTaskInput checks that a saved area task reflects the user input.
func tasksContainsTaskInput(tasksMD, input string) bool {
	input = strings.TrimSpace(strings.ReplaceAll(txt.NormNewLines(input), "\n", " "))
	if input == "" {
		return tasksMD == "" || !strings.Contains(tasksMD, "- [ ]")
	}
	return chatContainsTaskInput(tasksMD, input)
}

// chatContainsTaskInput checks that a saved inbox entry reflects the user input.
// Task text may differ by first-rune casing or invalid UTF-8 from fuzz inputs.
func chatContainsTaskInput(chat, input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return chat == ""
	}
	firstLine := strings.SplitN(input, "\n", 2)[0]
	if strings.Contains(chat, firstLine) {
		return true
	}
	if strings.Contains(chat, txt.Ucfirst(firstLine)) {
		return true
	}
	if len(firstLine) > 1 && strings.Contains(chat, firstLine[1:]) {
		return true
	}
	return strings.Contains(chat, "- [ ]")
}

func moveToKeyboard(h string) *tg.Keyboard {
	return tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrToTomorrow), tg.NewCmd(CmdScheduleForTmrw, []string{h})),
			tg.NewBtn(i18n.Tr(i18n.StrToLater), tg.NewCmd(CmdMoveToLater, []string{h})),
			tg.NewBtn(i18n.Tr(i18n.StrToADay), tg.NewCmd(CmdShowScheduleForDay, []string{h})),
		),
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrToFile), tg.NewCmd(CmdShowMoveToDirOrFile, []string{h})),
			tg.NewBtn(i18n.Tr(i18n.StrToJournal), tg.NewCmd(CmdMoveToJournal, []string{h})),
		),
	})
}
