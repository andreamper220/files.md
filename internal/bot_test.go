package internal

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"zakirullin/stuffbot/internal/userconfig"

	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/pkg/tg"
	"zakirullin/stuffbot/pkg/tg/fake"
)

func init() {
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}
}

func TestSaveFromTextMsg(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, "New task"))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(tasks, 1)
	r.Equal("New task.md", tasks[0].Name)

	content, err := bot.fs.Read("today", "New task.md")
	r.NoError(err)
	r.Empty(content)
}

func TestSaveFromLongTextMsg(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, strings.Repeat("a", 101)))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	filename := fmt.Sprintf("A%s....md", strings.Repeat("a", 99))
	r.Len(tasks, 1)
	r.Equal(filename, tasks[0].Name)

	content, err := bot.fs.Read("today", filename)
	r.NoError(err)
	r.Equal("A"+strings.Repeat("a", 100), content)
}

func TestSaveFromTextMsgWithSanitize(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, "New task/"))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(tasks, 1)
	r.Equal("New task{|}.md", tasks[0].Name)

	content, err := bot.fs.Read("today", "New task{|}.md")
	r.NoError(err)
	r.Equal("New task/", content)

	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("today", nil)))
	r.NoError(err)

	r.Equal("<b>1</b> left"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("👀 New task/", tg.NewCmd("task", []string{"today", "cd59b9e6546"})),
	},
	), tgram.SentKeyboard)
}

func TestAddMultilineTaskToToday(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, "New task\nContent"))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(tasks, 1)
	r.Equal("New task.md", tasks[0].Name)
	r.True(tasks[0].IsMultiline)

	content, err := bot.fs.Read("today", "New task.md")
	r.NoError(err)
	r.Equal("New task\nContent", content)
}

func TestAddTaskWithSpecCharsToToday(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, "New task\nUrl! http://g.com (Also_text] ##header\n-item1\n-item2\n1+1=2"))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(tasks, 1)
	r.Equal("New task.md", tasks[0].Name)
	r.True(tasks[0].IsMultiline)

	content, err := bot.fs.Read("today", "New task.md")
	r.NoError(err)
	r.Equal("New task\nUrl! http://g.com (Also\\_text] ##header\n-item1\n-item2\n1+1=2", content)
}

func TestSaveFromRegularReply(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, 8, 11, 9, 54, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Existing file.md", "Existing content")
	r.NoError(err)

	tgram := fake.NewTG()
	database := db.NewFakeDB()
	database.SetDirByMsgID(-1, 255, "today")
	database.SetFilenameByMsgID(-1, 255, "Existing file.md")
	bot := NewBot(-1, tgram, userFS, database, fakeConfig())

	upd := fake.NewUpd(-1, "Line")
	upd.ReplyToMessageID = 255
	err = bot.Answer(upd)
	r.NoError(err)

	files, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)
	r.Len(files, 1)

	content, err := bot.fs.Read("today", "Existing file.md")
	r.NoError(err)
	r.Equal("### 11.08.2024 Sunday\nLine\nExisting content", content)
}

func TestSaveFromPhotoWithCaption(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	upd.PhotoCaption = "Caption"
	err = bot.Answer(upd)
	r.NoError(err)

	files, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(files, 1)
	r.Equal("Caption.md", files[0].Name)
	r.True(files[0].IsMultiline)

	content, err := bot.fs.Read("today", "Caption.md")
	r.NoError(err)
	r.Equal("![[../img/tg_PHOTO_ID|center|400]]\nCaption", content)
}

func TestSaveFromPhotoWithLongCaption(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	upd.PhotoCaption = strings.Repeat("a", 101)
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := bot.fs.Read("today", fmt.Sprintf("A%s....md", strings.Repeat("a", 99)))
	r.NoError(err)
	r.Equal(fmt.Sprintf("![[../img/tg_PHOTO_ID|center|400]]\nA%s", strings.Repeat("a", 100)), content)
}

func TestSaveFromPhotoWithSanitizedCaption(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	upd.PhotoCaption = "Caption/"
	err = bot.Answer(upd)
	r.NoError(err)

	files, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(files, 1)
	r.Equal("Caption{|}.md", files[0].Name)
	r.True(files[0].IsMultiline)

	content, err := bot.fs.Read("today", "Caption{|}.md")
	r.NoError(err)
	r.Equal("![[../img/tg_PHOTO_ID|center|400]]\nCaption/", content)
}

func TestSaveFromPhotoWithoutCaption(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, 8, 11, 9, 54, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	err = bot.Answer(upd)
	r.NoError(err)

	files, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(files, 1)
	r.Equal("Img 11.08.24 09꞉54.md", files[0].Name)
	r.True(files[0].IsMultiline)

	content, err := bot.fs.Read("today", "Img 11.08.24 09꞉54.md")
	r.NoError(err)
	r.Equal("![[../img/tg_PHOTO_ID|center|400]]", content)
}

func TestSaveFromReplyPhotoWithCaption(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, 8, 11, 9, 54, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Existing file.md", "Existing content")
	r.NoError(err)

	tgram := fake.NewTG()

	database := db.NewFakeDB()
	database.SetDirByMsgID(-1, 255, "today")
	database.SetFilenameByMsgID(-1, 255, "Existing file.md")
	bot := NewBot(-1, tgram, userFS, database, fakeConfig())

	upd := fake.NewUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	upd.PhotoCaption = "Caption"
	upd.ReplyToMessageID = 255
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := bot.fs.Read("today", "Existing file.md")
	r.NoError(err)
	r.Equal("### 11.08.2024 Sunday\n![[../img/tg_PHOTO_ID|center|400]]\nCaption\nExisting content", content)
}

func TestAddTaskToLater(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.CreateDirsIfNotExist())

	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("mv", []string{"later", "today", "0824149b387"})))
	r.NoError(err)

	todayTasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)
	r.Len(todayTasks, 0)

	laterTasks, err := bot.fs.FilesAndDirs("later")
	r.NoError(err)
	r.Len(laterTasks, 1)
	r.Equal("First task.md", laterTasks[0].Name)
}

func TestCompleteTask(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("c", []string{"today", "0824149b387"})))
	r.NoError(err)

	todayTasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)
	r.Len(todayTasks, 0)

	completedTasks, err := bot.fs.FilesAndDirs("archive")
	r.NoError(err)
	r.Len(completedTasks, 1)
	r.Equal("First task.md", completedTasks[0].Name)
}

func TestToday(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	defer func() {
		fs.Ctime = savedCtime
	}()
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)
	err = userFS.Write("today", "Second task.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("today", nil)))
	r.NoError(err)

	r.Equal("<b>2</b> left"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("First task", tg.NewCmd("c", []string{"today", "0824149b387"})),
		tg.NewBtn("🥈 Second task", tg.NewCmd("c", []string{"today", "4eb62f93b3e"})),
	},
	), tgram.SentKeyboard)
}

func TestLater(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	defer func() {
		fs.Ctime = savedCtime
	}()
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("later", "First task.md", "")
	r.NoError(err)
	err = userFS.Write("later", "Second task", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("later", nil)))
	r.NoError(err)

	r.Equal("⏳ Your tasks for <b>later</b>:", tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("First task", tg.NewCmd("c", []string{"later", "0824149b387"})),
		tg.NewBtn("🥈 Second task", tg.NewCmd("c", []string{"later", "2940ad40402"})),
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	},
	), tgram.SentKeyboard)
}

func TestTodayQuickMenuFilled(t *testing.T) {
	savedCtime := fs.Ctime
	defer func() {
		fs.Ctime = savedCtime
	}()
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	cfg := fakeConfig()
	cfg.AddQuickCmd("files")
	cfg.AddQuickCmd("checklists")
	cfg.AddQuickCmd("postpone")
	bot, tgram, r := makeBot(t, cfg)
	err := bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("today", nil)))
	r.NoError(err)
	r.Equal("<b>1</b> left"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("First task", tg.NewCmd("c", []string{"today", "0824149b387"})),
		tg.NewRow(
			tg.NewBtn("📄", tg.NewCmd("files", nil)),
			tg.NewBtn("☑️", tg.NewCmd("checklists", nil)),
			tg.NewBtn("🦥", tg.NewCmd("postpone", nil)),
		),
	},
	), tgram.SentKeyboard)
}

func TestTodayWithMultilineTasks(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	defer func() {
		fs.Ctime = savedCtime
	}()
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "First task.md", "content")
	r.NoError(err)
	err = userFS.Write("today", "Second task.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	upd := fake.NewUpdCmdFake(-1, tg.NewCmd("today", nil))
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(upd)
	r.NoError(err)

	r.Equal("<b>2</b> left"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("👀 First task", tg.NewCmd("task", []string{"today", "0824149b387"})),
		tg.NewBtn("🥈 Second task", tg.NewCmd("c", []string{"today", "4eb62f93b3e"})),
	},
	), tgram.SentKeyboard)
}

func TestFiles(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("", "Doc1.md", "")
	r.NoError(err)
	err = userFS.Write("", "Doc2.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("files", nil)))
	r.NoError(err)

	r.Equal("📄 Your files:"+wideSpacer, tgram.SentTexts[0])
	r.Equal(tg.NewKeyboard([]tg.Row{
		[]tg.Btn{
			tg.NewBtn("📄 Doc1", tg.NewCmd("file", []string{fs.DirRoot, "c1253521ac7"})),
			tg.NewBtn("📄 Doc2", tg.NewCmd("file", []string{fs.DirRoot, "64572c3093f"})),
		},
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	}), tgram.SentKeyboard)
}

func TestChecklists(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("-checklist1-")
	r.NoError(err)
	err = userFS.MakeDir("-checklist2-")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("checklists", nil)))
	r.NoError(err)

	r.Equal("☑️ Checklists", tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("Checklist1", tg.NewCmd("checklist", []string{"8d2335b5ff3"})),
		tg.NewBtn("Checklist2", tg.NewCmd("checklist", []string{"8d3625e2e84"})),
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	},
	), tgram.SentKeyboard)
}

func TestAddSingleItemToChecklist(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("-checklist1-")
	r.NoError(err)
	err = userFS.Write("today", "Item.md", "")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("mv_to_chk", []string{"7b72407ca70", "-checklist1-"})))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("-checklist1-")
	r.NoError(err)
	r.Len(files, 1)
	r.Equal("Item.md", files[0].Name)

	files, err = userFS.FilesAndDirs("today")
	r.NoError(err)
	r.Len(files, 0)
}

func TestAddMultipleItemsToChecklist(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("-checklist1-")
	r.NoError(err)
	err = userFS.Write("today", "Item.md", "item\nitem2\nitem3\n\n")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("mv_to_chk", []string{"7b72407ca70", "-checklist1-"})))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("-checklist1-")
	r.NoError(err)
	r.Len(files, 3)
	r.ElementsMatch([]string{"Item.md", "Item2.md", "Item3.md"}, []string{files[0].Name, files[1].Name, files[2].Name})
}

func TestShowChecklist(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("-checklist1-")
	r.NoError(err)
	err = userFS.Write("-checklist1-", "Item.md", "")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("checklist", []string{"8d2335b5ff3"})))
	r.NoError(err)

	r.Equal("Checklist1"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("Item", tg.NewCmd("check_comp", []string{"8d2335b5ff3", "7b72407ca70"})),
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	},
	), tgram.SentKeyboard)
}

func TestCompleteItemInChecklist(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.MakeDir("-checklist1-")
	r.NoError(err)
	err = userFS.Write("-checklist1-", "Item.md", "")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("check_comp", []string{"8d2335b5ff3", "7b72407ca70"})))
	r.NoError(err)

	r.Equal("Checklist1"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	},
	), tgram.SentKeyboard)

	items, err := bot.fs.FilesAndDirs("-checklist1-")
	r.NoError(err)
	r.Empty(items)

	items, err = bot.fs.FilesAndDirs("archive")
	r.NoError(err)
	r.Len(items, 1)
	r.Equal("Item.md", items[0].Name)
}

func TestBotTodayLabelIcons(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	tgram := fake.NewTG()
	b := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	// Pomodoro is the only task in today
	r.Nil(b.togglePomodoro(nil))
	label := b.todayLabel()
	r.Contains(label, "🌴")
	r.Contains(label, "🍅")

	// Pomodoro and another task in today
	r.Nil(b.fs.Write(fs.DirToday, "Item.md", ""))
	label = b.todayLabel()
	r.NotContains(label, "🌴")
	r.Contains(label, "🍅")

	// No pomodoro, but there is another task in today
	r.Nil(b.complete([]string{fs.DirToday, fs.FilePomodoro}))
	label = b.todayLabel()
	r.NotContains(label, "🌴")
	r.NotContains(label, "🍅")

	// No pomodoro, no other tasks in today
	r.Nil(b.complete([]string{fs.DirToday, "Item.md"}))
	label = b.todayLabel()
	r.NoError(err)
	r.Contains(label, "🌴")
	r.NotContains(label, "🍅")
}

func makeBot(t *testing.T, conf ConfigInterface) (*Bot, *fake.TG, *require.Assertions) {
	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)
	err = userFS.Write("later", "Second task", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), conf)
	return bot, tgram, r
}

func TestSettingsMainPanel(t *testing.T) {
	bot, tgram, r := makeBot(t, fakeConfig())
	err := bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("settings", nil)))
	r.NoError(err)
	r.Equal("Settings:", tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("⚡️ Quick buttons", tg.NewCmd("c_quick_btns", nil)),
		tg.NewBtn("➡️ Move to buttons", tg.NewCmd("c_move_btns", nil)),
		tg.NewBtn("🏠 Today", tg.NewCmd("today", nil)),
	},
	), tgram.SentKeyboard)
}

// Quick Panel Data-driven tests

var (
	btnFilesDel      = tg.NewBtn("📄 Files ➖", tg.NewCmd("del_quick", []string{"files"}))
	btnChecklistsDel = tg.NewBtn("☑️ Checklists ➖", tg.NewCmd("del_quick", []string{"checklists"}))
	btnPostponeDel   = tg.NewBtn("🦥 Postpone ➖", tg.NewCmd("del_quick", []string{"postpone"}))
)

var (
	delimiter = tg.NewBtn("-", tg.NewCmd("nothing", nil))
	backBtn   = tg.NewBtn("⬅️ Back", tg.NewCmd("settings", nil))
)

var (
	btnLater          = tg.NewBtn("⏳ Later ➕", tg.NewCmd("add_quick", []string{"later"}))
	btnSearch         = tg.NewBtn("🔎 Search ➕", tg.NewCmd("add_quick", []string{"search"}))
	btnFilesAdd       = tg.NewBtn("📄 Files ➕", tg.NewCmd("add_quick", []string{"files"}))
	btnChecklistsAdd  = tg.NewBtn("☑️ Checklists ➕", tg.NewCmd("add_quick", []string{"checklists"}))
	btnPostponeAdd    = tg.NewBtn("🦥 Postpone ➕", tg.NewCmd("add_quick", []string{"postpone"}))
	btnReadChecklist  = tg.NewBtn("📚 Read ➕", tg.NewCmd("add_quick", []string{"read"}))
	btnWatchChecklist = tg.NewBtn("📺 Watch ➕", tg.NewCmd("add_quick", []string{"watch"}))
	btnShopChecklist  = tg.NewBtn("🛒 Shop ➕", tg.NewCmd("add_quick", []string{"shop"}))
	btnHabits         = tg.NewBtn("🌱 Habits ➕", tg.NewCmd("add_quick", []string{"habits"}))
)

func TestConfigureQP_Empty_Default(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{""},
		fake.NewUpdCmdFake(-1, tg.NewCmd("c_quick_btns", nil)),
		[]tg.Row{
			delimiter,
			btnLater,
			btnSearch,
			btnFilesAdd,
			btnChecklistsAdd,
			btnPostponeAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Empty_AddFiles(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{""},
		fake.NewUpdCmdFake(-1, tg.NewCmd("add_quick", []string{"files"})),
		[]tg.Row{
			btnFilesDel,
			delimiter,
			btnLater,
			btnSearch,
			btnChecklistsAdd,
			btnPostponeAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Doc_AddCheckList(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("add_quick", []string{"checklists"})),
		[]tg.Row{
			btnFilesDel,
			btnChecklistsDel,
			delimiter,
			btnLater,
			btnSearch,
			btnPostponeAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklists_AddPostpone(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("add_quick", []string{"postpone"})),
		[]tg.Row{
			btnFilesDel,
			btnChecklistsDel,
			btnPostponeDel,
			delimiter,
			btnLater,
			btnSearch,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklistsPostpone_Show(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists", "postpone"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("c_quick_btns", nil)),
		[]tg.Row{
			btnFilesDel,
			btnChecklistsDel,
			btnPostponeDel,
			delimiter,
			btnLater,
			btnSearch,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklistsPostpone_DelChecklists(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists", "postpone"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("del_quick", []string{"checklists"})),
		[]tg.Row{
			btnFilesDel,
			btnPostponeDel,
			delimiter,
			btnLater,
			btnSearch,
			btnChecklistsAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocPostpone_DelDoc(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "postpone"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("del_quick", []string{"files"})),
		[]tg.Row{
			btnPostponeDel,
			delimiter,
			btnLater,
			btnSearch,
			btnFilesAdd,
			btnChecklistsAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Postpone_DelPostpone(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"postpone"},
		fake.NewUpdCmdFake(-1, tg.NewCmd("del_quick", []string{"postpone"})),
		[]tg.Row{
			delimiter,
			btnLater,
			btnSearch,
			btnFilesAdd,
			btnChecklistsAdd,
			btnPostponeAdd,
			btnReadChecklist,
			btnWatchChecklist,
			btnShopChecklist,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Empty_AddUnknown(t *testing.T) {
	RunQuickPanelTc_Error(PrefTableTestCase{
		[]string{""},
		fake.NewUpdCmdFake(-1, tg.NewCmd("add_quick", []string{"wrong"})),
		[]tg.Row{},
	}, "unknown command: wrong", t)
}

func RunQuickPanelTc(tc PrefTableTestCase, t *testing.T) {
	cnf := fakeConfig()
	for _, cmd := range tc.existingCmds {
		cnf.AddQuickCmd(cmd)
	}

	bot, tgram, r := makeBot(t, cnf)

	err := bot.Answer(tc.updToAnswer)
	r.NoError(err)
	r.Equal("Configure quick buttons (➕ = add to quick buttons, ➖ = to remove from quick buttons):", tgram.LastSentText)
	r.Equal(tg.NewKeyboard(tc.buttons), tgram.SentKeyboard)
}

func RunQuickPanelTc_Error(tc PrefTableTestCase, expectedErr string, t *testing.T) {
	cnf := fakeConfig()
	for _, cmd := range tc.existingCmds {
		cnf.AddQuickCmd(cmd)
	}
	bot, _, r := makeBot(t, cnf)
	actualErr := bot.Answer(tc.updToAnswer)
	r.EqualError(actualErr, expectedErr)
}

type PrefTableTestCase struct {
	existingCmds []string
	updToAnswer  *fake.Upd
	buttons      []tg.Row
}

func TestShowToFileNoDirs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Note.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.showMoveToFile([]string{"345fbd7ab08"})
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn("📄 Note", tg.NewCmd("mf", []string{"345fb", "", "345fbd7ab08"}))),
	},
	), tgram.SentKeyboard)
}

func TestShowMoveToFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Note.md", "")
	r.NoError(err)
	err = userFS.MakeDir("dir")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.showMoveToFile([]string{"345fbd7ab08"})
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn("🗂️ dir", tg.NewCmd("mv", []string{"73600", "", "345fbd7ab08"}))),
		tg.NewBtn("Or choose a file:", tg.NewCmd("nothing", nil)),
		tg.NewRow(tg.NewBtn("📄 Note", tg.NewCmd("mf", []string{"345fb", "", "345fbd7ab08"}))),
	},
	), tgram.SentKeyboard)
}

func TestShow(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show("text", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Equal("text", tgram.LastSentText)
}

func TestShowLongMessage(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4096)+"b", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 2)
	r.Equal("b", tgram.LastSentText)
}

// When utf8.RuneCountInString(textChunk) == 4096, tg sends the message (len(textChunk) => 7003)
// if I have 4095 chars and add 🟢, we have 4096 chars and it is ok
// if I have 4095 chars and add ⚪️, we have 4097 chars and we fail, so tg doesn't operate on glyth clusters
func TestShowLongMessageWithColoredEmojis(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4095)+"🟢", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 1)
}

func TestShowLongMessageWithColoredEmoji(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4095)+"⚪️", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 2)
}

func TestShowLongMessageSplitByNewLine(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4094)+"\nabc", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 2)
	r.Equal("abc", tgram.LastSentText)
}

func TestShowMultilineFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write("today", "New file.md", "New file\nContent")

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("task", []string{fs.DirToday, "501ef2410e2"})))
	r.NoError(err)

	r.Equal("New file\nContent", tgram.SentTexts[0])
}

func TestMoveToExistingFile(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, 8, 11, 9, 54, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Task.md", "")
	r.NoError(err)
	err = userFS.Write("", "New file.md", "")
	r.NoError(err)
	err = userFS.Write("", "Existing file.md", "")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpdCmdFake(-1, tg.NewCmd("mf", []string{"1c8f819d075", "", "501ef2410e2"}))
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := userFS.Read("", "Existing file.md")
	r.NoError(err)
	r.Equal("### 11.08.2024 Sunday\nNew file", content)
}

func TestMoveToExistingFileExistingRecord(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, 8, 11, 9, 54, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Task.md", "")
	r.NoError(err)
	err = userFS.Write("", "New file.md", "")
	r.NoError(err)
	err = userFS.Write("", "Existing file.md", "### 11.08.2024 Sunday\nContent")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpdCmdFake(-1, tg.NewCmd("mf", []string{"1c8f819d075", "", "501ef2410e2"}))
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := userFS.Read("", "Existing file.md")
	r.NoError(err)
	r.Equal("### 11.08.2024 Sunday\nNew file\nContent", content)
}

func TestShowMoveTo(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("", "file", "")

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpd(-1, "New task\nContent"))
	r.NoError(err)

	r.Equal("Task added for <b>today</b>!", tgram.SentTexts[0])

	kb := tg.NewKeyboard([]tg.Row{
		[]tg.Btn{
			{Name: "🌚 To tmrw", Cmd: tg.Cmd{Name: "sc_tmrw", Params: []string{"d0776a3e2b9"}, Type: "cmd"}},
			{Name: "📆 To a day", Cmd: tg.Cmd{Name: "sc_day", Params: []string{"d0776a3e2b9"}, Type: "cmd"}},
			{Name: "📄 To File", Cmd: tg.Cmd{Name: "to_file", Params: []string{"d0776a3e2b9"}, Type: "cmd"}},
		},
		[]tg.Btn{
			{Name: "💚 To Journal", Cmd: tg.Cmd{Name: "mv_to_journal", Params: []string{"d0776a3e2b9"}, Type: "cmd"}},
			{Name: "➡️ Today", Cmd: tg.Cmd{Name: "today", Params: []string(nil), Type: "cmd"}},
		}},
	)
	r.Equal(kb, tgram.SentKeyboard)
}

func TestShowScheduleEmpty(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("schedule", nil)))
	r.NoError(err)

	r.Equal("You don't have any scheduled tasks! 🌴", tgram.SentTexts[0])
}

func TestShowSchedule(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := fake.NewTG()

	conf := fakeConfig()
	conf.AddToSchedule("filename.md", 0, "")
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), conf)
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("schedule", nil)))
	r.NoError(err)

	r.Equal("<b>01 January, Thursday</b>\n- Filename", tgram.SentTexts[0])
}

func TestAngerEmoji(t *testing.T) {
	r := require.New(t)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	file := fs.NewFile("n", "h", "t", 0, false, false, "")
	r.Equal("", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 1, 23, 59, 59, 999999999, time.UTC)
	}
	r.Equal("", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)
	}
	r.Equal("🙄", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 6, 0, 0, 0, 0, time.UTC)
	}
	r.Equal("🤬️", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 7, 0, 0, 0, 0, time.UTC)
	}
	r.Equal("🤬️", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)
	}
	file = fs.NewFile("n", "h", "t", 24*60*60, false, false, "")
	r.Equal("", angerEmoji(file))

	now = func() time.Time {
		return time.Date(1970, 1, 3, 0, 0, 0, 0, time.UTC)
	}
	file = fs.NewFile("n", "h", "t", 24*60*60, false, false, "")
	r.Equal("🙄", angerEmoji(file))
}

func TestAngerInTodayTasks(t *testing.T) {
	r := require.New(t)

	savedCtime := fs.Ctime
	defer func() {
		fs.Ctime = savedCtime
	}()
	fs.Ctime = func(fi os.FileInfo) int64 {
		return 0
	}

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Angry task.md", "")
	r.NoError(err)

	tgram := fake.NewTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(fake.NewUpdCmdFake(-1, tg.NewCmd("today", nil)))
	r.NoError(err)

	r.Equal("<b>1</b> left"+wideSpacer, tgram.LastSentText)
	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewBtn("🙄 Angry task", tg.NewCmd("c", []string{"today", "9c556351f34"})),
	},
	), tgram.SentKeyboard)
}

func TestMoveToChecklistSplittable(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Item1.md", "Item1\nItem2")
	r.NoError(err)
	err = userFS.MakeDir("-checklist-")
	r.NoError(err)

	tgram := fake.NewTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.moveToChecklist([]string{"Item1.md", "-checklist-"})
	r.NoError(err)

	files, err := userFS.FilesAndDirs("-checklist-")
	r.NoError(err)

	items := fs.OnlyFilenames(files)
	r.ElementsMatch([]string{"Item1.md", "Item2.md"}, items)
}

func fakeConfig() ConfigInterface {
	userFS, _ := fs.NewFS("/-1", afero.NewMemMapFs())
	cfg := userconfig.NewConfig(userFS, -1, "config.json")
	cfg.CreateDefaultIfNotExists()

	return cfg
}

func TestExtractCmd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, fake.NewTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "/t task for tomorrow")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("t", cmd.Name)
	r.Equal([]string{"Task for tomorrow"}, cmd.Params)
}

func TestExtractCmdRu(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, fake.NewTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "/т task for tomorrow")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("т", cmd.Name)
	r.Equal([]string{"Task for tomorrow"}, cmd.Params)
}

func TestExtractCmdSkips(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, fake.NewTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "/today task for tomorrow")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.Nil(cmd)
}

func TestExtractCmdAtTheEnd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, fake.NewTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := fake.NewUpd(-1, "task for tomorrow /t")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("t", cmd.Name)
	r.Equal([]string{"Task for tomorrow"}, cmd.Params)
}
