package internal

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"zakirullin/stuffbot/internal/journal"
	"zakirullin/stuffbot/internal/sched"
	"zakirullin/stuffbot/internal/userconfig"

	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/pkg/tg"
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "New task"))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, strings.Repeat("a", 101)))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "New task/"))
	r.NoError(err)

	tasks, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)

	r.Len(tasks, 1)
	r.Equal("New task{|}.md", tasks[0].Name)

	content, err := bot.fs.Read("today", "New task{|}.md")
	r.NoError(err)
	r.Equal("New task/", content)

	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("today", nil)))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "New task\nContent"))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "New task\nUrl! http://g.com (Also_text] ##header\n-item1\n-item2\n1+1=2"))
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

	tgram := tg.NewFakeTG()
	database := db.NewFakeDB()
	database.SetDirByMsgID(-1, 255, "today")
	database.SetFilenameByMsgID(-1, 255, "Existing file.md")
	bot := NewBot(-1, tgram, userFS, database, fakeConfig())

	upd := tg.NewFakeUpd(-1, "Line")
	upd.ReplyToMessageID = 255
	err = bot.Answer(upd)
	r.NoError(err)

	files, err := bot.fs.FilesAndDirs("today")
	r.NoError(err)
	r.Len(files, 1)

	content, err := bot.fs.Read("today", "Existing file.md")
	r.NoError(err)
	r.Equal("#### 11 August, Sunday\nLine\nExisting content", content)
}

func TestSaveFromPhotoWithCaption(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "")
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "")
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "")
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "")
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

	tgram := tg.NewFakeTG()

	database := db.NewFakeDB()
	database.SetDirByMsgID(-1, 255, "today")
	database.SetFilenameByMsgID(-1, 255, "Existing file.md")
	bot := NewBot(-1, tgram, userFS, database, fakeConfig())

	upd := tg.NewFakeUpd(-1, "")
	upd.PhotoID = "PHOTO_ID"
	upd.PhotoCaption = "Caption"
	upd.ReplyToMessageID = 255
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := bot.fs.Read("today", "Existing file.md")
	r.NoError(err)
	r.Equal("#### 11 August, Sunday\n![[../img/tg_PHOTO_ID|center|400]]\nCaption\nExisting content", content)
}

func TestAddTaskToLater(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	r.NoError(userFS.CreateDirsIfNotExist())

	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("mv", []string{"later", "today", "0824149b387"})))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("c", []string{"today", "0824149b387"})))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("today", nil)))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("later", nil)))
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
	err := bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("today", nil)))
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

	tgram := tg.NewFakeTG()

	upd := tg.NewFakeUpdCmd(-1, tg.NewCmd("today", nil))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("files", nil)))
	r.NoError(err)

	r.Equal("📄 Your files:"+wideSpacer, tgram.SentTexts[0])
	r.Equal(tg.NewKeyboard([]tg.Row{
		[]tg.Btn{
			tg.NewBtn("📄 Doc1", tg.NewCmd("file", []string{fs.DirRoot, "c1253521ac7"})),
			tg.NewBtn("📄 Doc2", tg.NewCmd("file", []string{fs.DirRoot, "64572c3093f"})),
		},
		tg.NewBtn("🔎 Search", tg.NewCustomCmd("search", nil, "iq")),
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("checklists", nil)))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("mv_to_chk", []string{"7b72407ca70", "-checklist1-"})))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("mv_to_chk", []string{"7b72407ca70", "-checklist1-"})))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("checklist", []string{"8d2335b5ff3"})))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("check_comp", []string{"8d2335b5ff3", "7b72407ca70"})))
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
	tgram := tg.NewFakeTG()
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

func makeBot(t *testing.T, cfg *userconfig.Config) (*Bot, *tg.FakeTG, *require.Assertions) {
	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "First task.md", "")
	r.NoError(err)
	err = userFS.Write("later", "Second task", "")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), cfg)
	return bot, tgram, r
}

func TestSettingsMainPanel(t *testing.T) {
	bot, tgram, r := makeBot(t, fakeConfig())
	err := bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("settings", nil)))
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
	btnSchedule       = tg.NewBtn("🗓 Schedule ➕", tg.NewCmd("add_quick", []string{"schedule"}))
	btnHabits         = tg.NewBtn("🌱 Habits ➕", tg.NewCmd("add_quick", []string{"habits"}))
)

func TestConfigureQP_Empty_Default(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{""},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("c_quick_btns", nil)),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Empty_AddFiles(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{""},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("add_quick", []string{"files"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Doc_AddCheckList(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("add_quick", []string{"checklists"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklists_AddPostpone(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("add_quick", []string{"postpone"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklistsPostpone_Show(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists", "postpone"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("c_quick_btns", nil)),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocChecklistsPostpone_DelChecklists(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "checklists", "postpone"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("del_quick", []string{"checklists"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_DocPostpone_DelDoc(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"files", "postpone"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("del_quick", []string{"files"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Postpone_DelPostpone(t *testing.T) {
	RunQuickPanelTc(PrefTableTestCase{
		[]string{"postpone"},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("del_quick", []string{"postpone"})),
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
			btnSchedule,
			btnHabits,
			backBtn,
		},
	}, t)
}

func TestConfigureQP_Empty_AddUnknown(t *testing.T) {
	RunQuickPanelTc_Error(PrefTableTestCase{
		[]string{""},
		tg.NewFakeUpdCmd(-1, tg.NewCmd("add_quick", []string{"wrong"})),
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
	updToAnswer  *tg.FakeUpd
	buttons      []tg.Row
}

func TestShowToFileNoDirs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("today", "Note.md", "")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.showMoveToFileOrDir([]string{"345fbd7ab08"})
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn("📄 Note", tg.NewCmd("mf", []string{"345fb", "", "345fb"}))),
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.showMoveToFileOrDir([]string{"345fbd7ab08"})
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn("📄 Note", tg.NewCmd("mf", []string{"345fb", "", "345fb"}))),
		tg.NewBtn("Or choose a dir:", tg.NewCmd("nothing", nil)),
		tg.NewRow(tg.NewBtn("🗂️ dir", tg.NewCmd("mv", []string{"73600", "", "345fbd7ab08"}))),
	},
	), tgram.SentKeyboard)
}

func TestShow(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show("text", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Equal("text", tgram.LastSentText)
}

func TestShowLongMessage(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := tg.NewFakeTG()

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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4095)+"🟢", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 1)
}

func TestShowLongMessageWithColoredEmoji(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.show(strings.Repeat("a", 4095)+"⚪️", nil, tg.MarkupHTML)
	r.NoError(err)

	r.Len(tgram.SentTexts, 2)
}

func TestShowLongMessageSplitByNewLine(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	tgram := tg.NewFakeTG()

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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("task", []string{fs.DirToday, "501ef2410e2"})))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpdCmd(-1, tg.NewCmd("mf", []string{"1c8f819d075", "", "501ef2410e2"}))
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := userFS.Read("", "Existing file.md")
	r.NoError(err)
	r.Equal("#### 11 August, Sunday\nNew file", content)
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpdCmd(-1, tg.NewCmd("mf", []string{"1c8f819d075", "", "501ef2410e2"}))
	err = bot.Answer(upd)
	r.NoError(err)

	content, err := userFS.Read("", "Existing file.md")
	r.NoError(err)
	r.Equal("#### 11 August, Sunday\nNew file\n### 11.08.2024 Sunday\nContent", content)
}

func TestShowMoveTo(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.Write("", "file", "")

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "New task\nContent"))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("schedule", nil)))
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

	tgram := tg.NewFakeTG()

	cfg := fakeConfig()
	cfg.AddToSchedule("filename.md", 0, "")
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), cfg)
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("schedule", nil)))
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

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("today", nil)))
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

	tgram := tg.NewFakeTG()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.moveToChecklist([]string{"Item1.md", "-checklist-"})
	r.NoError(err)

	files, err := userFS.FilesAndDirs("-checklist-")
	r.NoError(err)

	items := fs.OnlyFilenames(files)
	r.ElementsMatch([]string{"Item1.md", "Item2.md"}, items)
}

func fakeConfig() *userconfig.Config {
	userFS, _ := fs.NewFS("/-1", afero.NewMemMapFs())
	cfg := userconfig.NewConfig(userFS, -1, "config.json")
	cfg.CreateDefaultIfNotExists()

	return cfg
}

func TestExtractCmd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "jj journal record")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("j", cmd.Name)
	r.Equal([]string{"Journal record"}, cmd.Params)
}

func TestExtractCmdRu(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "жж запись в журнал")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("j", cmd.Name)
	r.Equal([]string{"Запись в журнал"}, cmd.Params)
}

func TestExtractCmdSkipsInTheBeginning(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "jjj task for tomorrow")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.Nil(cmd)
}

func TestExtractCmdSkipsAtTheMiddle(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "journal jj record")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)
	r.Nil(cmd)
}

func TestExtractCmdSkipsInTheEnd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "task for tomorrow jjj")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)
	r.Nil(cmd)
}

func TestExtractCmdAtTheEnd(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)

	bot := NewBot(-1, tg.NewFakeTG(), userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "journal record jj")
	cmd, err := bot.extractCmd(upd)
	r.NoError(err)

	r.NotNil(cmd)
	r.Equal("j", cmd.Name)
	r.Equal([]string{"Journal record"}, cmd.Params)
}

func TestMoveToJournal(t *testing.T) {
	r := require.New(t)

	savedNow := journal.Now
	defer func() {
		now = savedNow
	}()
	journal.Now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	userFS.Write("today", "Note.md", "Multiline\ncontent")

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("mv_to_journal", []string{"345fbd7ab08"})))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("journal")
	r.NoError(err)
	r.Len(files, 1)

	content, err := userFS.Read("journal", files[0].Name)
	r.NoError(err)
	r.Equal("#### 1 January, Thursday\n`00:00` Multiline\ncontent\n", content)
}

func TestAddToJournalFromShortcut(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "jj record"))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("journal")
	r.NoError(err)
	r.Len(files, 1)
}

func TestAddToJournalFromShortcutRu(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "/ж запись"))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("journal")
	r.NoError(err)
	r.Len(files, 1)
}

func TestAddToJournalFromShortcutRuCases(t *testing.T) {
	r := require.New(t)

	savedNow := journal.Now
	defer func() {
		now = savedNow
	}()
	journal.Now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpd(-1, "жЖ запись"))
	r.NoError(err)

	files, err := userFS.FilesAndDirs("journal")
	r.NoError(err)
	r.Len(files, 1)

	content, err := userFS.Read("journal", files[0].Name)
	r.NoError(err)
	r.Equal("#### 1 January, Thursday\n`00:00` Запись\n", content)

	err = bot.Answer(tg.NewFakeUpd(-1, "Запись2 ЖЖ"))
	r.NoError(err)

	content, err = userFS.Read("journal", files[0].Name)
	r.NoError(err)
	r.Equal("#### 1 January, Thursday\n`00:00` Запись\n`00:00` Запись2\n", content)
}

func TestShowForADay(t *testing.T) {
	r := require.New(t)

	savedNow := sched.Now
	defer func() {
		now = savedNow
	}()
	sched.Now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("sc_day", []string{"1c8f819d075"})))
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{[]tg.Btn{tg.Btn{Name: "🔄️ Repeat the task", Cmd: tg.Cmd{Name: "sc_day_r", Params: []string{"1c8f819d075"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "Mon", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "345600", ""}, Type: "cmd"}}, tg.Btn{Name: "Tue", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "432000", ""}, Type: "cmd"}}, tg.Btn{Name: "Wed", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "518400", ""}, Type: "cmd"}}, tg.Btn{Name: "Thu", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "604800", ""}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "Fri", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "691200", ""}, Type: "cmd"}}, tg.Btn{Name: "Sat", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "172800", ""}, Type: "cmd"}}, tg.Btn{Name: "Sun", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "259200", ""}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "1", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2678400", ""}, Type: "cmd"}}, tg.Btn{Name: "2", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2764800", ""}, Type: "cmd"}}, tg.Btn{Name: "3", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "172800", ""}, Type: "cmd"}}, tg.Btn{Name: "4", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "259200", ""}, Type: "cmd"}}, tg.Btn{Name: "5", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "345600", ""}, Type: "cmd"}}, tg.Btn{Name: "6", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "432000", ""}, Type: "cmd"}}, tg.Btn{Name: "7", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "518400", ""}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "9", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "691200", ""}, Type: "cmd"}}, tg.Btn{Name: "10", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "777600", ""}, Type: "cmd"}}, tg.Btn{Name: "11", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "864000", ""}, Type: "cmd"}}, tg.Btn{Name: "12", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "950400", ""}, Type: "cmd"}}, tg.Btn{Name: "13", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1036800", ""}, Type: "cmd"}}, tg.Btn{Name: "14", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1123200", ""}, Type: "cmd"}}, tg.Btn{Name: "15", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1209600", ""}, Type: "cmd"}}, tg.Btn{Name: "16", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1296000", ""}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "17", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1382400", ""}, Type: "cmd"}}, tg.Btn{Name: "18", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1468800", ""}, Type: "cmd"}}, tg.Btn{Name: "19", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1555200", ""}, Type: "cmd"}}, tg.Btn{Name: "20", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1641600", ""}, Type: "cmd"}}, tg.Btn{Name: "21", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1728000", ""}, Type: "cmd"}}, tg.Btn{Name: "22", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1814400", ""}, Type: "cmd"}}, tg.Btn{Name: "23", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1900800", ""}, Type: "cmd"}}, tg.Btn{Name: "24", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "1987200", ""}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "25", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2073600", ""}, Type: "cmd"}}, tg.Btn{Name: "26", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2160000", ""}, Type: "cmd"}}, tg.Btn{Name: "27", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2246400", ""}, Type: "cmd"}}, tg.Btn{Name: "28", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2332800", ""}, Type: "cmd"}}, tg.Btn{Name: "29", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2419200", ""}, Type: "cmd"}}, tg.Btn{Name: "30", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2505600", ""}, Type: "cmd"}}, tg.Btn{Name: "31", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f819d075", "2592000", ""}, Type: "cmd"}}}, tg.Btn{Name: "➡️ Move to Today", Cmd: tg.Cmd{Name: "today", Params: []string(nil), Type: "cmd"}}}), tgram.SentKeyboard)
}

func TestShowForADayRecurring(t *testing.T) {
	r := require.New(t)

	savedNow := sched.Now
	defer func() {
		now = savedNow
	}()
	sched.Now = func() time.Time {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("sc_day_r", []string{"1c8f819d075"})))
	r.NoError(err)

	r.Equal(tg.NewKeyboard([]tg.Row{[]tg.Btn{tg.Btn{Name: "Weekdays", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "345600", "0 0 * * 1-5"}, Type: "cmd"}}, tg.Btn{Name: "Every day", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "172800", "0 0 * * *"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "Mon", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "345600", "0 0 * * 1"}, Type: "cmd"}}, tg.Btn{Name: "Tue", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "432000", "0 0 * * 2"}, Type: "cmd"}}, tg.Btn{Name: "Wed", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "518400", "0 0 * * 3"}, Type: "cmd"}}, tg.Btn{Name: "Thu", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "604800", "0 0 * * 4"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "Fri", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "691200", "0 0 * * 5"}, Type: "cmd"}}, tg.Btn{Name: "Sat", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "172800", "0 0 * * 6"}, Type: "cmd"}}, tg.Btn{Name: "Sun", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "259200", "0 0 * * 0"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "1", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2678400", "0 0 1 * *"}, Type: "cmd"}}, tg.Btn{Name: "2", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2764800", "0 0 2 * *"}, Type: "cmd"}}, tg.Btn{Name: "3", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "172800", "0 0 3 * *"}, Type: "cmd"}}, tg.Btn{Name: "4", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "259200", "0 0 4 * *"}, Type: "cmd"}}, tg.Btn{Name: "5", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "345600", "0 0 5 * *"}, Type: "cmd"}}, tg.Btn{Name: "6", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "432000", "0 0 6 * *"}, Type: "cmd"}}, tg.Btn{Name: "7", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "518400", "0 0 7 * *"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "8", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "604800", "0 0 8 * *"}, Type: "cmd"}}, tg.Btn{Name: "9", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "691200", "0 0 9 * *"}, Type: "cmd"}}, tg.Btn{Name: "10", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "777600", "0 0 10 * *"}, Type: "cmd"}}, tg.Btn{Name: "11", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "864000", "0 0 11 * *"}, Type: "cmd"}}, tg.Btn{Name: "12", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "950400", "0 0 12 * *"}, Type: "cmd"}}, tg.Btn{Name: "13", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1036800", "0 0 13 * *"}, Type: "cmd"}}, tg.Btn{Name: "14", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1123200", "0 0 14 * *"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "15", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1209600", "0 0 15 * *"}, Type: "cmd"}}, tg.Btn{Name: "16", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1296000", "0 0 16 * *"}, Type: "cmd"}}, tg.Btn{Name: "17", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1382400", "0 0 17 * *"}, Type: "cmd"}}, tg.Btn{Name: "18", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1468800", "0 0 18 * *"}, Type: "cmd"}}, tg.Btn{Name: "19", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1555200", "0 0 19 * *"}, Type: "cmd"}}, tg.Btn{Name: "20", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1641600", "0 0 20 * *"}, Type: "cmd"}}, tg.Btn{Name: "21", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1728000", "0 0 21 * *"}, Type: "cmd"}}}, []tg.Btn{tg.Btn{Name: "22", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1814400", "0 0 22 * *"}, Type: "cmd"}}, tg.Btn{Name: "23", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1900800", "0 0 23 * *"}, Type: "cmd"}}, tg.Btn{Name: "24", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "1987200", "0 0 24 * *"}, Type: "cmd"}}, tg.Btn{Name: "25", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2073600", "0 0 25 * *"}, Type: "cmd"}}, tg.Btn{Name: "26", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2160000", "0 0 26 * *"}, Type: "cmd"}}, tg.Btn{Name: "27", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2246400", "0 0 27 * *"}, Type: "cmd"}}, tg.Btn{Name: "28", Cmd: tg.Cmd{Name: "sc", Params: []string{"1c8f", "2332800", "0 0 28 * *"}, Type: "cmd"}}}, tg.Btn{Name: "➡️ Move to Today", Cmd: tg.Cmd{Name: "today", Params: []string(nil), Type: "cmd"}}}), tgram.SentKeyboard)
}

func TestSchedule(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("today", "Task.md", "")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	cfg := fakeConfig()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), cfg)
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("sc", []string{"8a42", "345600", "0 0 * * 1-5"})))
	r.NoError(err)

	tasksForToday, err := userFS.FilesAndDirs("today")
	r.NoError(err)
	r.Empty(tasksForToday)

	tasksForLater, err := userFS.FilesAndDirs("later")
	r.NoError(err)
	r.Len(tasksForLater, 1)
	r.Equal("Task.md", tasksForLater[0].Name)

	sc, err := cfg.Schedules()
	r.NoError(err)
	r.Len(sc, 1)
	r.Equal("Task.md", sc[0].Filename)
	r.Equal(int64(345600), sc[0].ScheduledAt)
	r.Equal("0 0 * * 1-5", sc[0].Cron)
}

func TestScheduleNoRepeat(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("today", "Task.md", "")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	cfg := fakeConfig()
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), cfg)
	err = bot.Answer(tg.NewFakeUpdCmd(-1, tg.NewCmd("sc", []string{"8a42", "345600", ""})))
	r.NoError(err)

	tasksForToday, err := userFS.FilesAndDirs("today")
	r.NoError(err)
	r.Empty(tasksForToday)

	tasksForLater, err := userFS.FilesAndDirs("later")
	r.NoError(err)
	r.Len(tasksForLater, 1)
	r.Equal("Task.md", tasksForLater[0].Name)

	sc, err := cfg.Schedules()
	r.NoError(err)
	r.Len(sc, 1)
	r.Equal("Task.md", sc[0].Filename)
	r.Equal(int64(345600), sc[0].ScheduledAt)
}

func TestInlineRequestTask(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("today", "Task.md", "Task content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, " today / Task ")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.NoError(err)

	r.Equal("Task\nTask content", tgram.LastSentText)
}

func TestInlineRequestFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, " File ")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.NoError(err)

	r.Equal("File\nFile content", tgram.LastSentText)
}

func TestInlineRequestFileOutsideTheDirectory(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "../File")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.Error(err)
	r.EqualError(err, "insecure input '../File': invalid request from inline query")
}

func TestInlineRequestFileOutsideTheDirectoryTwoSlashes(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "..//File")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.Error(err)
	r.EqualError(err, "insecure input '..//File': invalid request from inline query")
}

func TestInlineRequestFileListOutsideDirs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "../")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.Error(err)
	r.EqualError(err, "insecure input '../': invalid request from inline query")
}

func TestInlineRequestFileListRootDirs(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	// cd /tmp//.. would lead to root
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "/..")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.Error(err)
	r.EqualError(err, "insecure input '/..': invalid request from inline query")
}

func TestInlineRequestFileListRootDirsWithoutSlash(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	// cd /tmp/.. would lead to root
	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())
	upd := tg.NewFakeUpd(-1, "..")
	upd.IsSentViaBotVal = true

	err = bot.Answer(upd)
	r.Error(err)
	r.EqualError(err, "show file: can't unhash '..' in '': cannot unhash, maybe the file is missing")
}

func TestAnswerSearch(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	u := tg.NewFakeUpd(-1, "")
	u.InlineQueryVal = "File"
	u.IsInlineQueryVal = true

	err = bot.answerSearch(u)
	r.NoError(err)

	var results []interface{}
	article := tgbotapi.NewInlineQueryResultArticleHTML("0", "File", "File.md")
	results = append(results, article)

	r.Equal(results, tgram.InlineQueryResults)
}

func TestAnswerSearchShowAllRoot(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)
	err = userFS.MakeDir("Dir")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	u := tg.NewFakeUpd(-1, "")
	u.InlineQueryVal = " "
	u.IsInlineQueryVal = true

	err = bot.answerSearch(u)
	r.NoError(err)

	var results []interface{}
	article := tgbotapi.NewInlineQueryResultArticleHTML("0", "File", "File.md")
	results = append(results, article)

	r.Equal(results, tgram.InlineQueryResults)
}

func TestAnswerSearchShowOutsideTheRoot(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)
	err = userFS.MakeDir("Dir")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	u := tg.NewFakeUpd(-1, "")
	u.InlineQueryVal = "../"
	u.IsInlineQueryVal = true

	err = bot.answerSearch(u)
	r.Error(err)
	r.EqualError(err, "insecure input '../': invalid inline query")
}

func TestAnswerSearchShowOutsideTheRootNoSlash(t *testing.T) {
	r := require.New(t)

	memFS := afero.NewMemMapFs()
	err := afero.WriteFile(memFS, "/secret", []byte("secret"), 0o644)
	r.NoError(err)

	userFS, err := fs.NewFS("/-1", memFS)
	r.NoError(err)
	err = userFS.CreateDirsIfNotExist()
	r.NoError(err)
	err = userFS.Write("", "File.md", "File content")
	r.NoError(err)
	err = userFS.MakeDir("Dir")
	r.NoError(err)

	tgram := tg.NewFakeTG()

	bot := NewBot(-1, tgram, userFS, db.NewFakeDB(), fakeConfig())

	u := tg.NewFakeUpd(-1, "")
	u.InlineQueryVal = ".."
	u.IsInlineQueryVal = true

	err = bot.answerSearch(u)
	r.Error(err)
	r.EqualError(err, "inline reply: search notes: exists: unsafe path '/': unsafe path, possible security issue")
}
