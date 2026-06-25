// Bot's main functionality. We accept messages from the user,
// we ask user where to save the messages. We save messages
// to plain markdown files locally.

package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slog"

	"github.com/zakirullin/files.md/server/config"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/habits"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/journal"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/morningsummary"
	"github.com/zakirullin/files.md/server/pkg/slice"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/plugins"
	"github.com/zakirullin/files.md/server/priority"
	"github.com/zakirullin/files.md/server/stats"
	"github.com/zakirullin/files.md/server/stt"
	"github.com/zakirullin/files.md/server/sync"
	"github.com/zakirullin/files.md/server/userconfig"
)

var (
	errUnknownCommand           = errors.New("unknown command")
	errInvalidRequestFromInline = errors.New("invalid request from inline query")
	errInvalidInlineQuery       = errors.New("invalid inline query")
	BotPlugins                  = []BotPlugin{plugins.NewWorldClockPlugin()}
)

const (
	btnsPerRow               = 3
	quickBtnsPerRow          = 4
	maxBtns                  = 50
	maxBtnsInChecklist       = 10 // For _read_ and _watch_ checklists, so we're less likely to be overwhelmed :)
	maxGroupedBtnsInMoveTo   = 6
	maxInlineResults         = 20
	maxMsgLength             = 4096 // In UTF-8 characters (runes), skin-tone emojis count as 2
	maxMsgsToSendAtOnce      = 5    // For lengthy messages
	maxHeaderLength          = 100
	maxHeaderLengthForMobile = 18 // Task button preview in tasks list
	inlineResultsCacheTime   = 15 // Seconds

	// On mobile phones buttons shrink to the message width, and sometimes it's too narrow, so we make the message wider
	wideSpacer = "<code>            ⁠</code>"
)

// Update represents incoming user updates.
type Update interface {
	MsgText() string
	UserID() int64
	Cmd() *tg.Cmd
	MsgEntities() []tgbotapi.MessageEntity
	CaptionEntities() []tgbotapi.MessageEntity
	CallbackQueryID() (string, bool)
	InlineQueryID() (string, bool)
	InlineQuery() (string, bool)
	InlineQueryOffset() int
	IsSentViaBot() bool
	ReplyToMsgID() (int, bool)
	PhotoOrImageID() (string, bool)
	AudioOnlyID() (string, bool)
	DocumentOnlyID() (string, bool)
	MediaGroupID() (string, bool)
	DocumentFileName() string
	Caption() string
	MsgID() (int, bool)
	Time() (int, bool)
	ChannelID() (int64, bool)
	ChannelName() (string, bool)
}

// Chat provides a simple interface to chat API like Telegram.
type Chat interface {
	Send(userID int64, text string, kb *tg.Keyboard, markup string) (int, error)
	SendImages(userID int64, images []string) ([]int, error)
	SendReaction(userID int64, msgID int, reaction string) error
	Edit(userID int64, msgID int, text string, kb *tg.Keyboard, markup string) error
	Del(userID int64, msgID int) error
	AnswerCallbackQuery(queryID string, text string) error
	AnswerInlineQuery(queryID string, results []interface{}, cacheTime int, offset string) error
	DownloadFile(fileID string, outFile io.Writer) (string, error)
	SendDocument(userID int64, filename string, content io.Reader, caption string, kb *tg.Keyboard) (int, error)
}

// Database stores per user data like "last sent message id"
type Database interface {
	LastKeyboardMsgID() (int, bool)
	SetLastKeyboardMsgID(ID int)
	DelLastKeyboardMsgID()
	InputExpectation() *tg.Cmd
	SetInputExpectation(cmd tg.Cmd)
	DelInputExpectation()
	HashOrPathByMsgID(msgID int) (string, bool)
	SetHashOrPathByMsgID(msgID int, value string)
	RecentCommand() (string, bool)
	SetRecentCommand(cmd string)
	RecentCommandParams() ([]string, bool)
	SetRecentCommandParams(params []string)
	AddImgMsgID(msgID int)
	ImgMsgID() ([]int, bool)
	DelImgMsgID()
	SetPendingDraft(hash, content string)
	PendingDraft(hash string) (content string, ok bool)
	DelPendingDraft(hash string)
}

type BotPlugin interface {
	CanHandle(string) bool
	Handle(string) (output string, error error)
}

var now = time.Now

// Telegram only allows 64 bytes in callback_data,
// So we have to be really short :)
const (
	CmdShowStart                       = "start"
	CmdDoNothing                       = "nothing"
	CmdShowLater                       = "later"
	CmdShowHome                        = "home"
	CmdShowFiles                       = "files"
	CmdShowDirs                        = "dirs"
	CmdShowPostpone                    = "postpone"
	CmdShowMoveExisting                = "move"
	CmdShowMoveTo                      = "s_move"
	CmdShowRename                      = "rename"
	CmdShowRenameFile                  = "rename_file"
	CmdShowChecklists                  = "checklists"
	CmdShowStats                       = "stats"
	CmdOpenInApp                       = "app"
	CmdShowHelp                        = "help"
	CmdComplete                        = "c"
	CmdPostpone                        = "post"
	CmdShowLongItemFromChecklist       = "item"
	CmdShowLongItem                    = "item_t"
	CmdShowFile                        = "file"
	CmdShowChecklist                   = "checklist"
	CmdShowChecklistItem               = "check_show"
	CmdCompleteListItem                = "check_comp"
	CmdShowMoveToDirOrFile             = "to_file"
	CmdShowMoveToChecklist             = "to_checklist"
	CmdRename                          = "ren"
	CmdMoveToExistingDir               = "mv"
	CmdMoveToChecklist                 = "add_item"
	CmdCompleteChecklistItem           = "check_item"
	CmdRequestNewDir                   = "new_dir"
	CmdMoveToNewDir                    = "mv_to_new_dir"
	CmdMoveToExistingFile              = "mf"
	CmdMoveToExistingNote              = "mvn"
	CmdMoveToNewFile                   = "mn"
	CmdMoveToDirChecklist              = "mv_to_chk"
	CmdMoveToRead                      = "mv_to_read"
	CmdMoveToWatch                     = "mv_to_watch"
	CmdMoveToShop                      = "mv_to_shop"
	CmdMoveToNewChecklist              = "mv_to_new_chk"
	CmdMoveToJournal                   = "mv_to_journal"
	CmdMoveToLater                     = "mv_later"
	CmdShowScheduleForDay              = "sc_day"
	CmdSchedule                        = "sc"
	CmdScheduleForTmrw                 = "sc_tmrw"
	CmdPomodoro                        = "pomodoro"
	CmdShowScheduleForDayRecurring     = "sc_day_r"
	CmdLater                           = "later"
	CmdShowSettings                    = "settings"
	CmdShowQuickBtnsSettings           = "c_quick_btns"
	CmdShowMoveToBtnsSettings          = "c_move_btns"
	CmdAddToQuickBtns                  = "add_quick"
	CmdDelFromQuickBtns                = "del_quick"
	CmdAddToMoveToBtns                 = "add_move"
	CmdDelFromMoveToBtns               = "del_move"
	CmdShowTimezone                    = "timezone"
	CmdSetTimezone                     = "set_timezone"
	CmdShowReadChecklist               = "read"
	CmdShowWatchChecklist              = "watch"
	CmdShowShopChecklist               = "shop"
	CmdShowSchedule                    = "schedule"
	CmdDownload                        = "download"
	CmdTasksOnlyMode                   = "tasks_only"
	CmdNotesOnlyMode                   = "notes_only"
	CmdJournalOnlyMode                 = "journal_only"
	CmdFullMode                        = "full"
	CmdChatMode                        = "chat"
	CmdInlineQuerySearchEveryWhere     = "search"
	CmdWebAppHabits                    = "habits"
	CmdRandomNote                      = "random_note"
	CmdAddToJournalShortcut            = "j"
	CmdAddToJournalAndContinueShortcut = "ja"
	CmdAddToRecentFileShortcut         = "+"
	CmdCompleteHabit                   = "ch"
	CmdShare                           = "share"
	CmdSetPriority                     = "prio"
	CmdShowMorning                     = "morning"
	CmdInitLife                        = "life_init"
	CmdShowLifeSpheres                 = "life_sph"
	CmdShowLifeSphere                  = "life_s1"
	CmdShowLifeProject                 = "life_p1"
	CmdShowLifeDocs                    = "life_doc"
	CmdMoveToDraft                     = "mv_draft"
	CmdMoveToFinalize                  = "mv_fin"
	CmdMoveToDiscussion                = "mv_disc"
	CmdFinalizeDoc                     = "finalize"
	CmdLifePickProject                 = "lp_p"
	CmdLifeSaveToProject               = "lp_sv"
	CmdLifeCreateProject               = "lp_cr"
	CmdLifeNewProject                  = "lp_new"
	CmdLifeCreateProjectOnly           = "lp_cro"
	CmdLifeNewSection                  = "lp_nsec"
	CmdLifeCreateSection               = "lp_csec"
	CmdShowMoveToSphere                = "mv_sph"
	CmdMoveToSphere                    = "mv_sp"
	CmdAddToDraftShortcut              = "draft_sc"
	CmdAddToFinalizeShortcut           = "fin_sc"
	CmdAddToDiscussionShortcut         = "disc_sc"
	CmdShowSaveType                    = "save_type"
	CmdCancelPendingDraft              = "draft_x"
	CmdSaveAsTask                      = "as_task"
	CmdPickTaskPriority                = "task_prio"
	CmdSaveAsNote                      = "as_note"
	CmdApplyDraftTitle                 = "d_title"
	CmdShowTasksView                   = "tasks_v"
	CmdShowNotesHub                    = "notes_h"
	CmdSaveNoteToArea                  = "note_area"
	CmdSaveTaskToArea                  = "task_area"
	CmdMoveToAreaTask                  = "mv_area"
	CmdMoveAreaTask                    = "mv_ar"
	CmdMoveNoteKind                    = "note_k"
	CmdShowMoveNoteArea                = "note_ma"
	CmdMoveNoteArea                    = "note_m2"
	CmdCompleteAreaTask                = "area_c"
	CmdShowTask                        = "task_show"
	CmdDeleteTask                      = "task_del"
	CmdOpenMedia                       = "open_m"
	CmdShowTaskActions               = "task_act"
)

var Shortcuts = map[string][]string{
	CmdAddToJournalShortcut:            {"/ж", "jj", "жж"},
	CmdAddToJournalAndContinueShortcut: {"жд", "jd", "ja"},
	CmdAddToRecentFileShortcut:         {"++"},
	CmdAddToDraftShortcut:              {"/д", "draft", "черновик", "д"},
	CmdAddToFinalizeShortcut:           {"/ф", "fin", "фин", "финализировать", "ф"},
	CmdAddToDiscussionShortcut:         {"/об", "disc", "обсуждение", "об"},
}

var allowedSlashCommands = map[string]bool{
	CmdShowHome:  true,
	CmdShowDirs:  true,
	CmdShowFiles: true,
}

// Bot has all the things that we need to handle a message or command from a user.
// We use tg chat to talk with the user.
// We use fs to save artefacts to the disk (.md files).
// We use db to save temporal things like recent command.
// We use cfg to configure bot behaviour (config.json).
type Bot struct {
	userID int64
	tg     Chat
	fs     *fs.FS
	db     Database
	cfg    *userconfig.Config
}

func NewBot(userID int64, tg Chat, fs *fs.FS, db Database, cfg *userconfig.Config) *Bot {
	return &Bot{userID, tg, fs, db, cfg}
}

// Reply to incoming text message, command or inline query
func (b *Bot) Reply(u Update) error {
	// Handle inline queries.
	if _, ok := u.InlineQueryID(); ok {
		return b.answerSearch(u)
	}

	// Handle messages from channels.
	_, isChannel := u.ChannelID()
	if isChannel {
		channelName, _ := u.ChannelName()
		if len(strings.TrimSpace(channelName)) == 0 {
			channelName = "UnknownChannel"
		}

		return b.addToFile(fs.DirUserRoot, fs.Filename(channelName), u.MsgText())
	}

	// Handle plugins.
	for _, plugin := range BotPlugins {
		if plugin.CanHandle(u.MsgText()) {
			output, err := plugin.Handle(u.MsgText())
			if err != nil {
				return fmt.Errorf("answer: plugin error: %w", err)
			}
			_, _ = b.tg.Send(b.userID, output, nil, tg.MarkupHTML)

			b.delAllKeyboards()
			err = b.ShowHome(nil)
			if err != nil {
				return fmt.Errorf("answer after plugin: %w", err)
			}

			return nil
		}
	}

	// Handle inline query file requests
	if u.IsSentViaBot() {
		return b.answerFileRequest(u.MsgText())
	}

	// Handle commands
	cmd, err := b.extractCmd(u)
	if err != nil {
		return fmt.Errorf("answer: %w", err)
	}
	if cmd != nil {
		if _, ok := u.CallbackQueryID(); !ok {
			b.delAllKeyboards()
		}

		handler, ok := b.handlers()[cmd.Name]
		if !ok {
			// It should be handled at cmd extraction step
			return fmt.Errorf("no such command %s: %w", cmd.Name, errUnknownCommand)
		}
		slog.Debug("Command is called", "command", cmd.Name, "params", cmd.Params)
		err = handler(cmd.Params)
		if err != nil {
			return err
		}

		if callbackQueryID, ok := u.CallbackQueryID(); ok {
			// We can tolerate an error here, that won't affect UX
			if cmd.Name == CmdCompleteHabit || cmd.Name == CmdComplete || cmd.Name == CmdCompleteAreaTask {
				_ = b.tg.AnswerCallbackQuery(callbackQueryID, completedMsg())
			} else if cmd.Name == CmdShare {
				_ = b.tg.AnswerCallbackQuery(callbackQueryID, i18n.Tr("Shared 💚!"))
			} else {
				_ = b.tg.AnswerCallbackQuery(callbackQueryID, "")
			}
		}

		return nil
	}

	if _, hasAudio := u.AudioOnlyID(); hasAudio {
		err = b.saveFromAudio(u)
	} else if _, hasImage := u.PhotoOrImageID(); hasImage {
		err = b.saveFromImage(u)
	} else if _, hasDoc := u.DocumentOnlyID(); hasDoc {
		err = b.saveFromDocument(u)
	} else {
		err = b.saveFromTextMsg(u)
	}

	if errors.Is(err, fs.ErrQuotaExceeded) {
		b.tg.Send(b.userID, i18n.Tr("Storage quota exceeded. Please delete some files."), nil, tg.MarkupHTML)
		return nil
	}

	return err
}

// Commands and their handlers.
// Every handler accepts []string params
func (b *Bot) handlers() map[string]func([]string) error {
	handlers := map[string]func([]string) error{
		// Direct user commands
		CmdShowHome:           b.ShowHome,
		CmdShowStart:          b.showStart,
		CmdShowLater:          b.showLaterTasks,
		CmdShowFiles:          b.showFiles,
		CmdShowDirs:           b.showDirs,
		CmdShowChecklists:     b.showChecklists,
		CmdShowPostpone:       b.showPostpone,
		CmdShowMoveTo:         b.showMoveTo,
		CmdShowRename:         b.showRename,
		CmdShowStats:          b.showStats,
		CmdShowMorning:        b.showMorningSummary,
		CmdInitLife:            b.initLife,
		CmdShowLifeSpheres:     b.showLifeSpheres,
		CmdShowLifeSphere:      b.showLifeSphere,
		CmdShowLifeProject:     b.showLifeProject,
		CmdShowLifeDocs:        b.showLifeDocs,
		CmdMoveToDraft:         b.moveToDraft,
		CmdMoveToFinalize:      b.moveToFinalize,
		CmdMoveToDiscussion:    b.moveToDiscussion,
		CmdFinalizeDoc:         b.finalizeDoc,
		CmdLifePickProject:     b.showLifeProjectPicker,
		CmdLifeSaveToProject:   b.saveToLifeProject,
		CmdLifeCreateProject:   b.createLifeProject,
		CmdLifeNewProject:      b.lifeNewProject,
		CmdLifeCreateProjectOnly: b.createLifeProjectOnly,
		CmdLifeNewSection:        b.lifeNewSection,
		CmdLifeCreateSection:     b.createLifeSection,
		CmdShowMoveToSphere:    b.showMoveToSphere,
		CmdMoveToSphere:        b.moveToSphere,
		CmdAddToDraftShortcut:      b.addToDraftFromShortcut,
		CmdAddToFinalizeShortcut:   b.addToFinalizeFromShortcut,
		CmdAddToDiscussionShortcut: b.addToDiscussionFromShortcut,
		CmdSetPriority:               b.setPriority,
		CmdShowSaveType:              b.showSaveType,
		CmdCancelPendingDraft:        b.cancelPendingDraft,
		CmdSaveAsTask:                b.saveAsTask,
		CmdPickTaskPriority:          b.pickTaskPriority,
		CmdSaveAsNote:                b.saveAsNote,
		CmdApplyDraftTitle:           b.applyDraftTitle,
		CmdShowTasksView:             b.showTasksView,
		CmdShowNotesHub:              b.showNotesHub,
		CmdSaveNoteToArea:            b.saveNoteToArea,
		CmdSaveTaskToArea:            b.saveTaskToArea,
		CmdMoveToAreaTask:            b.showMoveAreaTask,
		CmdMoveAreaTask:              b.moveAreaTask,
		CmdMoveNoteKind:              b.moveNoteKind,
		CmdShowMoveNoteArea:          b.showMoveNoteArea,
		CmdMoveNoteArea:              b.moveNoteArea,
		CmdCompleteAreaTask:          b.completeAreaTask,
		CmdShowTask:                  b.showTask,
		CmdDeleteTask:                b.deleteTask,
		CmdOpenMedia:                 b.openMedia,
		CmdShowTaskActions:           b.showTaskActions,
		CmdRandomNote:                b.randomNote,
		CmdShowMoveExisting:   b.showMoveExisting,
		CmdShowSettings:       b.showSettings,
		CmdShowTimezone:       b.showTimezone,
		CmdSetTimezone:        b.setTimezone,
		CmdShowHelp:           b.showHelp,
		CmdDownload:           b.download,
		// Button's commands (callbacks)
		CmdShowRenameFile:                  b.showRenameFile,
		CmdShowLongItemFromChecklist:       b.showLongItemFromChecklist,
		CmdShowLongItem:                    b.showLongItem,
		CmdShowFile:                        b.showFile,
		CmdShowChecklist:                   b.showChecklist,
		CmdCompleteListItem:                b.completeListItem,
		CmdShowChecklistItem:               b.showChecklistItem,
		CmdShowScheduleForDay:              b.showToADay,
		CmdShowMoveToDirOrFile:             b.showMoveToFileOrDir,
		CmdShowMoveToChecklist:             b.showToChecklist,
		CmdMoveToExistingDir:               b.moveToDir,
		CmdMoveToChecklist:                 b.moveToChecklist,
		CmdCompleteChecklistItem:           b.completeChecklistItem,
		CmdRequestNewDir:                   b.requestNewDirName,
		CmdMoveToNewDir:                    b.moveToNewDir,
		CmdMoveToExistingFile:              b.moveToExistingFile,
		CmdMoveToExistingNote:              b.moveToExistingNote,
		CmdMoveToNewFile:                   b.moveToNewFile,
		CmdMoveToDirChecklist:              b.moveToDirChecklist,
		CmdMoveToRead:                      b.moveToRead,
		CmdMoveToWatch:                     b.moveToWatch,
		CmdMoveToShop:                      b.moveToShop,
		CmdMoveToNewChecklist:              b.moveToNewChecklist,
		CmdMoveToJournal:                   b.moveToJournal,
		CmdMoveToLater:                     b.moveToLater,
		CmdSchedule:                        b.schedule,
		CmdScheduleForTmrw:                 b.scheduleForTmrw,
		CmdComplete:                        b.complete,
		CmdPostpone:                        b.postpone,
		CmdPomodoro:                        b.togglePomodoro,
		CmdShowScheduleForDayRecurring:     b.showToADayRecurring,
		CmdShowQuickBtnsSettings:           b.showQuickBtnsSettings,
		CmdShowMoveToBtnsSettings:          b.showMoveToBtnsSettings,
		CmdAddToQuickBtns:                  b.addToQuickBtns,
		CmdDelFromQuickBtns:                b.delFromQuickBtns,
		CmdAddToMoveToBtns:                 b.addToMoveToBtns,
		CmdDelFromMoveToBtns:               b.delFromMoveToBtns,
		CmdAddToJournalShortcut:            b.addToJournalFromShortcut,
		CmdAddToJournalAndContinueShortcut: b.addToJournalAndContinue,
		CmdAddToRecentFileShortcut:         b.addToRecentFileOrNoteFromShortcut,
		CmdRename:                          b.rename,
		CmdTasksOnlyMode:                   b.setTasksOnlyMode,
		CmdNotesOnlyMode:                   b.setNotesOnlyMode,
		CmdJournalOnlyMode:                 b.setJournalOnlyMode,
		CmdFullMode:                        b.setFullMode,
		CmdChatMode:                        b.setChatOnlyMode,
		CmdCompleteHabit:                   b.completeHabit,
		CmdShare:                           b.shareNote,
		CmdShowDeleteFile:                  b.showDeleteFileConfirm,
		CmdDeleteFile:                      b.deleteFile,
		CmdShowDeleteDir:                   b.showDeleteDirConfirm,
		CmdDeleteDir:                       b.deleteDir,
		// Used for button-like separators
		CmdDoNothing: func(s []string) error { return nil },
	}

	for cmd, shortcuts := range Shortcuts {
		for _, shortcut := range shortcuts {
			handlers[shortcut] = handlers[cmd]
		}
	}

	return handlers
}

func (b *Bot) extractCmd(u Update) (*tg.Cmd, error) {
	cmd := u.Cmd()
	if cmd != nil {
		if _, isCallback := u.CallbackQueryID(); !isCallback && !allowedSlashCommands[cmd.Name] {
			_, _ = b.tg.Send(b.userID, i18n.Tr("I know nothing about this command 😕"), nil, tg.MarkupHTML)
			return nil, fmt.Errorf("unknown command: %s", cmd.Name)
		}

		// Check if the command is known
		_, ok := b.handlers()[cmd.Name]
		if !ok {
			_, _ = b.tg.Send(b.userID, i18n.Tr("I know nothing about this command 😕"), nil, tg.MarkupHTML)
			return nil, fmt.Errorf("unknown command: %s", cmd.Name)
		}

		b.db.DelInputExpectation()

		return cmd, nil
	}

	// Input expectation is mostly used for renaming things
	cmd = b.db.InputExpectation()
	if cmd != nil {
		if cmd.Name == CmdApplyDraftTitle && isMediaUpdate(u) {
			b.db.DelInputExpectation()
		} else {
			slog.Debug("Got command from input expectation", "command", cmd.Name)
			b.db.DelInputExpectation()

			for i, param := range cmd.Params {
				if param == "%s" {
					cmd.Params[i] = u.MsgText()
				}
			}

			return cmd, nil
		}
	}

	for canonicalCMD, shortcuts := range Shortcuts {
		for _, shortcut := range shortcuts {
			escapedShortcut := regexp.QuoteMeta(shortcut)
			reText := regexp.MustCompile(fmt.Sprintf(`(?i)^%s\s+|\s+%s$`, escapedShortcut, escapedShortcut))
			// The only difference from reText is that caption can contain only shortcut, with no other text
			reCaption := regexp.MustCompile(fmt.Sprintf(`(?i)^%s\s+|\s+%s$|^\s*%s\s*$`, escapedShortcut, escapedShortcut, escapedShortcut))

			doesntMatchText := !reText.MatchString(u.MsgText())
			doesntMatchCaption := !reCaption.MatchString(u.Caption())
			if doesntMatchText && doesntMatchCaption {
				continue
			}

			text := ""
			_, hasImage := u.PhotoOrImageID()
			if hasImage {
				var errImage error
				text, errImage = b.saveImage(u)
				if errImage != nil {
					return nil, fmt.Errorf("save image: %w", errImage)
				}
				text = string(reCaption.ReplaceAll([]byte(text), []byte("")))
			} else {
				text = extractMarkdown(u)
				text = string(reText.ReplaceAll([]byte(text), []byte("")))
			}

			text = txt.Ucfirst(strings.TrimSpace(text))
			shortCmd := tg.NewCmd(canonicalCMD, []string{text})

			return &shortCmd, nil
		}
	}

	return nil, nil
}

func isMediaUpdate(u Update) bool {
	if _, ok := u.AudioOnlyID(); ok {
		return true
	}
	if _, ok := u.PhotoOrImageID(); ok {
		return true
	}
	if _, ok := u.DocumentOnlyID(); ok {
		return true
	}
	return false
}

func (b *Bot) saveFromTextMsg(u Update) error {
	msg := extractMarkdown(u)
	if len(msg) == 0 {
		return fmt.Errorf("save: empty message")
	}

	// Collapse a few consecutive messages into one, see bot_forwards.go
	msgTime, updateHasTime := u.Time()
	if updateHasTime {
		_, shouldCollapse := collapseToMsg(b.userID, msgTime)
		if shouldCollapse {
			// We just write at the end of our append-only chat file,
			// that would concat the current message with the previous one.
			err := b.createOrAdd(fs.DirUserRoot, fs.ChatFilename, msg)
			if err != nil {
				return fmt.Errorf("save collapsed: %w", err)
			}
			return nil
		}
	}

	// Adding to an existing file or chat item
	if replyMsgID, ok := u.ReplyToMsgID(); ok {
		return b.addToReplied(replyMsgID, msg)
	}

	if updateHasTime {
		setFirstMsgTime(b.userID, msgTime)
	}

	return b.queueIncomingContent(msg)
}

// TODO test collapsing from both regular messages and images
func (b *Bot) saveFromImage(u Update) error {
	content, err := b.saveImage(u)
	if err != nil {
		return fmt.Errorf("save from image: %w", err)
	}

	// Collapse a few consecutive messages into one, see bot_forwards.go
	msgTime, updateHasTime := u.Time()
	if updateHasTime {
		_, shouldCollapse := collapseToMsg(b.userID, msgTime)
		if shouldCollapse {
			err := b.createOrAdd(fs.DirUserRoot, fs.ChatFilename, content)
			if err != nil {
				return fmt.Errorf("save collapsed: %w", err)
			}
			return nil
		}
	}

	// Adding to an existing file or chat item
	if replyMsgID, ok := u.ReplyToMsgID(); ok {
		return b.addToReplied(replyMsgID, content)
	}

	if b.cfg.ChatOnlyMode() {
		if updateHasTime {
			setFirstMsgTime(b.userID, msgTime)
		}
		return b.queueIncomingContent(content)
	}

	if updateHasTime {
		setFirstMsgTime(b.userID, msgTime)
	}

	return b.queueIncomingContent(content)
}

func (b *Bot) saveFromAudio(u Update) error {
	statusID := b.sendVoiceTranscriptionStatus()

	content, err := b.saveAudio(u)
	if statusID > 0 {
		b.finishVoiceTranscriptionStatus(statusID, content, err)
	}
	if err != nil {
		return fmt.Errorf("save from audio: %w", err)
	}

	if replyMsgID, ok := u.ReplyToMsgID(); ok {
		return b.addToReplied(replyMsgID, content)
	}

	return b.queueIncomingContent(content)
}

func (b *Bot) sendVoiceTranscriptionStatus() int {
	var msg string
	if config.ServerCfg.KieAPIKey != "" {
		msg = i18n.Tr("🎙 Расшифровываю…")
	} else {
		msg = i18n.Tr("🎙 Сохраняю голосовое…")
	}
	id, err := b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)
	if err != nil {
		return 0
	}
	return id
}

func (b *Bot) finishVoiceTranscriptionStatus(statusID int, content string, saveErr error) {
	var msg string
	switch {
	case saveErr != nil:
		msg = i18n.Tr("🎙 Ошибка сохранения")
	case config.ServerCfg.KieAPIKey == "":
		msg = i18n.Tr("🎙 Сохранено")
	case txt.VoiceSummary(content) != "":
		msg = i18n.Tr("🎙 Расшифровано ✓")
	default:
		msg = i18n.Tr("🎙 Не удалось расшифровать")
	}
	_ = b.tg.Edit(b.userID, statusID, msg, nil, tg.MarkupHTML)
	_ = b.tg.Del(b.userID, statusID)
}

func (b *Bot) saveAudio(u Update) (string, error) {
	audioID, _ := u.AudioOnlyID()

	var buf bytes.Buffer
	extension, err := b.tg.DownloadFile(audioID, &buf)
	if err != nil {
		return "", fmt.Errorf("can't download audio: %w", err)
	}

	audioFilename := fmt.Sprintf("tg_%s%s", audioID, extension)
	err = b.fs.Write(fs.DirMedia, audioFilename, buf.String())
	if err != nil {
		return "", fmt.Errorf("can't save audio: %w", err)
	}

	audioPath := fmt.Sprintf("%s/%s", fs.DirMedia, audioFilename)
	mimeType := "audio/ogg"
	switch strings.ToLower(extension) {
	case ".mp3", ".mpeg":
		mimeType = "audio/mpeg"
	case ".wav":
		mimeType = "audio/wav"
	case ".m4a", ".mp4":
		mimeType = "audio/mp4"
	}

	var content string
	if config.ServerCfg.KieAPIKey != "" {
		transcript, sttErr := stt.Transcribe(config.ServerCfg.KieAPIKey, buf.Bytes(), mimeType)
		if sttErr != nil {
			slog.Warn("voice transcription failed", "err", sttErr)
		} else if strings.TrimSpace(transcript) != "" {
			content = txt.FormatVoiceContent(strings.TrimSpace(transcript))
		}
	}
	if content == "" {
		content = txt.VoicePlaceholder
	}
	content = fmt.Sprintf("%s\n\n![](%s)", content, audioPath)

	if u.Caption() != "" {
		caption := txt.TelegramEntitiesToMarkdown(u.Caption(), u.CaptionEntities())
		content = fmt.Sprintf("%s\n%s", content, strings.TrimSpace(caption))
	}

	return content, nil
}

func (b *Bot) saveFromDocument(u Update) error {
	content, err := b.saveDocument(u)
	if err != nil {
		return fmt.Errorf("save from document: %w", err)
	}

	if replyMsgID, ok := u.ReplyToMsgID(); ok {
		return b.addToReplied(replyMsgID, content)
	}

	if groupID, ok := u.MediaGroupID(); ok {
		return b.bufferMediaGroupContent(groupID, content, u.Caption(), u.CaptionEntities())
	}

	return b.queueIncomingContent(content)
}

func mediaStorageFilename(fsys *fs.FS, originalName, docID, extension string) (string, error) {
	originalName = strings.TrimSpace(originalName)
	if originalName == "" {
		originalName = fmt.Sprintf("tg_%s%s", docID, extension)
	} else {
		originalName = filepath.Base(originalName)
		if extension != "" && !strings.HasSuffix(strings.ToLower(originalName), strings.ToLower(extension)) {
			originalName += extension
		}
	}
	return fsys.UniqueMediaFilename(originalName)
}

func (b *Bot) saveDocument(u Update) (string, error) {
	docID, _ := u.DocumentOnlyID()

	var buf bytes.Buffer
	extension, err := b.tg.DownloadFile(docID, &buf)
	if err != nil {
		return "", fmt.Errorf("can't download document: %w", err)
	}

	docFilename, err := mediaStorageFilename(b.fs, u.DocumentFileName(), docID, extension)
	if err != nil {
		return "", fmt.Errorf("can't pick document filename: %w", err)
	}
	if err := b.fs.Write(fs.DirMedia, docFilename, buf.String()); err != nil {
		return "", fmt.Errorf("can't save document: %w", err)
	}

	docPath := fmt.Sprintf("%s/%s", fs.DirMedia, docFilename)
	displayName := txt.AttachmentDisplayName("", docPath)
	if name := strings.TrimSpace(u.DocumentFileName()); name != "" {
		displayName = fs.SanitizeFilename(filepath.Base(name))
	}
	content := txt.FormatAttachmentContent(docPath, displayName)
	if u.Caption() != "" {
		if _, grouped := u.MediaGroupID(); !grouped {
			caption := txt.TelegramEntitiesToMarkdown(u.Caption(), u.CaptionEntities())
			content = fmt.Sprintf("%s\n%s", strings.TrimSpace(caption), content)
		}
	}

	return content, nil
}

// saveImage saves an image to the filesystem and returns a markdown link to it
func (b *Bot) saveImage(u Update) (string, error) {
	imageID, _ := u.PhotoOrImageID()

	var buf bytes.Buffer
	extension, err := b.tg.DownloadFile(imageID, &buf)
	if err != nil {
		return "", fmt.Errorf("can't download file: %w", err)
	}

	imgFilename := fmt.Sprintf("tg_%s%s", imageID, extension)
	err = b.fs.Write(fs.DirMedia, imgFilename, buf.String())
	if err != nil {
		return "", fmt.Errorf("can't save image: %w", err)
	}

	// TODO remove center
	imgPath := fmt.Sprintf("%s/%s", fs.DirMedia, imgFilename)
	content := fmt.Sprintf("![](%s)", imgPath)
	// If there's caption, place it under the image
	if u.Caption() != "" {
		caption := txt.TelegramEntitiesToMarkdown(u.Caption(), u.CaptionEntities())
		caption = strings.TrimSpace(txt.NormNewLines(caption))
		content = fmt.Sprintf("%s\n%s", content, txt.Ucfirst(caption))
	}

	return content, nil
}

// addToReplied appends newContent to whatever the bot rendered for
// replyToMsgID. Chat-item targets are stored with a "#" prefix
// ("#<msgHash>"); file targets are stored as a plain relative path.
func (b *Bot) addToReplied(replyToMsgID int, newContent string) error {
	value, ok := b.db.HashOrPathByMsgID(replyToMsgID)
	if !ok {
		return fmt.Errorf("add to replied: no target for msgID %d", replyToMsgID)
	}

	if strings.HasPrefix(value, "#") {
		msgHash := strings.TrimPrefix(value, "#")
		chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
		if err != nil {
			return fmt.Errorf("add to replied chat: can't read chat: %w", err)
		}
		updated, err := appendToChatMsg(chatMD, msgHash, newContent)
		if err != nil {
			return fmt.Errorf("add to replied chat: %w", err)
		}
		if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, updated); err != nil {
			return fmt.Errorf("add to replied chat: can't write chat: %w", err)
		}
		b.delAllKeyboards()
		return b.ShowHome(nil)
	}

	// File case: value is a relative path under user root.
	dir, filename := splitUserRelativePath(value)
	existingContent, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("add: can't read: %w", err)
	}

	content := strings.TrimRight(existingContent, "\n") + "\n\n" + strings.TrimSpace(newContent) + "\n"
	if err := b.fs.Write(dir, filename, content); err != nil {
		return fmt.Errorf("add: can't write: %w", err)
	}

	b.delAllKeyboards()

	b.db.SetRecentCommand(CmdMoveToExistingFile)
	b.db.SetRecentCommandParams([]string{fs.ShortHash(filepath.ToSlash(filepath.Join(dir, filename)))})

	return b.ShowHome(nil)
}

func (b *Bot) answerSearch(u Update) error {
	query, ok := u.InlineQuery()
	if !ok {
		return nil
	}
	query = strings.TrimSpace(query)

	if strings.Contains(query, "../") || strings.Contains(query, "/..") {
		return fmt.Errorf("insecure input '%s': %w", query, errInvalidInlineQuery)
	}

	matchedNotes, err := b.fs.SearchFilesByName(query)
	if err != nil {
		return fmt.Errorf("inline reply: %w", err)
	}
	if u.InlineQueryOffset() >= len(matchedNotes) {
		return nil
	}
	maxIndex := min(u.InlineQueryOffset()+maxInlineResults, len(matchedNotes))
	matchedNotes = matchedNotes[u.InlineQueryOffset():maxIndex]

	var results []interface{}
	for id, note := range matchedNotes {
		// Nested files: show the relative path without leading "/" so the
		// search result reads like "happiness/sub/Note.md".
		path := note.Name
		if note.ParentDir != fs.DirUserRoot && note.ParentDir != "" {
			path = fmt.Sprintf("<code>%s/%s</code>", note.ParentDir, note.Name)
		}
		article := tgbotapi.NewInlineQueryResultArticleHTML(strconv.Itoa(id), note.DisplayName, path)
		results = append(results, article)
	}

	queryID, _ := u.InlineQueryID()
	nextOffset := strconv.Itoa(u.InlineQueryOffset() + maxInlineResults)

	// First element is usually the file itself, exclude it
	if len(query) == 0 {
		results = results[1:]
	}

	err = b.tg.AnswerInlineQuery(queryID, results, inlineResultsCacheTime, nextOffset)
	// FakeTG library has a bug of unmarshalling sent result, we'll mute that temporarily
	if err != nil && !strings.HasSuffix(err.Error(), "Go value of type tgbotapi.Message") {
		return fmt.Errorf("inline reply: %w", err)
	}

	return nil
}

func (b *Bot) answerFileRequest(msg string) error {
	if strings.Contains(msg, "../") || strings.Contains(msg, "/..") {
		return fmt.Errorf("insecure input '%s': %w", msg, errInvalidRequestFromInline)
	}

	// Split on the FIRST slash: dir is the top-level directory (what
	// Unhash resolves against root), path is everything after it. Lets
	// nested inline-search results like
	// "triggers/habits/insights/2022 Habits.md" parse as
	// (dir="triggers", path="habits/insights/2022 Habits.md").
	msg = strings.TrimSpace(msg)
	var dir, path string
	if idx := strings.Index(msg, "/"); idx == -1 {
		dir = fs.DirUserRoot
		path = msg
	} else {
		dir = strings.TrimSpace(msg[:idx])
		path = strings.TrimSpace(msg[idx+1:])
	}
	if path == "" {
		return fmt.Errorf("invalid inline query '%s': %w", msg, errInvalidRequestFromInline)
	}

	b.delAllKeyboards()

	// TODO add tests
	// User wants to add his text to a selected file
	c := b.db.InputExpectation()
	if c != nil {
		b.db.DelInputExpectation()
		msgHash := c.Params[0]

		err := b.moveFromChat(func(content string, timestamp time.Time) error {
			if dir == fs.DirUserRoot {
				// We have a file
				b.db.SetRecentCommand(CmdMoveToExistingFile)
				b.db.SetRecentCommandParams([]string{fs.ShortHash(path)})
			} else {
				// We have a note (a file placed in a subdirectory)
				b.db.SetRecentCommand(CmdMoveToExistingNote)
				b.db.SetRecentCommandParams([]string{fs.ShortHash(path), fs.ShortHash(dir)})
			}

			err := b.addToFile(dir, path, content)
			if err != nil {
				return fmt.Errorf("inline query: can't add to file %s: %w", path, err)
			}

			return nil
		}, false, msgHash)
		if err != nil {
			return fmt.Errorf("inline query: can't move from chat: %w", err)
		}

		// Just an informative message
		_, _ = b.tg.Send(b.userID, fmt.Sprintf(i18n.Tr("Saved to <b>%s</b>"), fs.DisplayName(path)), nil, tg.MarkupHTML)

		return b.ShowHome(nil)
	}

	return b.showFile([]string{dir, path})
}

func (b *Bot) createOrAdd(dir, filename, content string) error {
	exists, err := b.fs.Exists(dir, filename)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	if exists {
		existingContent, err := b.fs.Read(dir, filename)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		existingContent = strings.TrimSpace(existingContent)

		if len(existingContent) != 0 {
			content = fmt.Sprintf("%s\n%s", strings.TrimSpace(existingContent), content)
		}
	}

	if err := b.fs.Write(dir, filename, content); err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (b *Bot) extractHeaderAndBody(msg string, maxHeaderLen int) (string, string, error) {
	if len(msg) == 0 {
		return "", "", fmt.Errorf("extract title: empty msg")
	}

	parts := strings.Split(msg, "\n")
	firstLine := strings.TrimSpace(parts[0])
	title := txt.Ucfirst(firstLine)
	if att, ok := txt.ParseAttachmentLine(firstLine); ok {
		if draftTitle := txt.DraftTitle(msg); draftTitle != "" {
			title = txt.Ucfirst(draftTitle)
		} else if names := txt.AttachmentNames(msg); len(names) > 0 {
			title = txt.Ucfirst(txt.AttachmentNoteTitle(names))
		} else {
			title = txt.Ucfirst(txt.AttachmentDisplayName(att.Name, att.Path))
		}
	}
	if title == txt.VoicePlaceholder {
		if alt := txt.VoiceSummary(msg); alt != "" {
			title = txt.Ucfirst(alt)
		}
	}
	if txt.HasImage(title) {
		if len(parts) > 1 {
			title = txt.Ucfirst(strings.TrimSpace(parts[1]))
		}

		if title == "" || len(parts) == 1 {
			title = fmt.Sprintf(i18n.Tr("Img %s"), now().Format("02.01.06 15:04"))
		}
	}

	if utf8.RuneCountInString(title) > maxHeaderLen {
		title = txt.Substr(title, 0, maxHeaderLen) + "..."
	}

	sanitizedTitle := fs.SanitizeFilename(title)
	content := msg
	// If title is the same as content, we don't need to save it
	if sanitizedTitle == content {
		content = ""
	}
	// If title is already in the content, remove it.
	// See bot.restoreMsg() to see how the message is restored.
	if strings.HasPrefix(content, sanitizedTitle) {
		content = strings.TrimSpace(strings.TrimPrefix(content, sanitizedTitle))
	}

	return sanitizedTitle, content, nil
}

// extractHeaderAndBodyPreserveMedia keeps image/audio attachments in the body.
func (b *Bot) extractHeaderAndBodyPreserveMedia(msg string, maxHeaderLen int) (string, string, error) {
	title, body, err := b.extractHeaderAndBody(msg, maxHeaderLen)
	if err != nil {
		return "", "", err
	}
	if txt.HasImage(msg) && !strings.Contains(body, "!(") {
		body = msg
	}
	if strings.TrimSpace(body) == "" {
		body = msg
	}
	return title, body, nil
}

// If content is empty, use its filename as content.
// If file has content, add filename to the beginning of the content.
// If file has content, and filename was truncated (...), no need to add filename.
// If file has image and caption underneath it, no need to add title.
// The ugliest method so far.
func (b *Bot) restoreMsg(dir, filename string) (string, error) {
	msg, err := b.fs.Read(dir, filename)
	if err != nil {
		return "", fmt.Errorf("can't restore msg for '%s': %w", filename, err)
	}

	title := fs.DisplayName(filename)
	nonTruncatedTitle := strings.TrimRight(title, "...")
	sanitizedContent := strings.ToLower(fs.SanitizeFilename(msg))
	contentHasNoTitle := !strings.HasPrefix(sanitizedContent, strings.ToLower(nonTruncatedTitle))
	hasNoImg := !txt.HasImage(msg)
	if len(msg) == 0 {
		return title, nil
	} else if contentHasNoTitle && hasNoImg {
		return fmt.Sprintf("%s\n%s", title, msg), nil
	}

	// msg has all the information, title doesn't have anything to add
	return msg, nil
}

func (b *Bot) tr(str string, args ...any) string {
	str = i18n.Tr(str)

	return fmt.Sprintf(str, args...)
}

// Replace last message + keyboard with the new one
// Or show the new one (in case of wimagehoto).
func (b *Bot) showHTML(validHTML string, kb *tg.Keyboard) error {
	b.delAllImages()

	mid, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if !hasLastKeyboard {
		b.delAllKeyboards()

		mid, err := b.tg.Send(b.userID, validHTML, kb, tg.MarkupHTML)
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}

		b.db.SetLastKeyboardMsgID(mid)

		return nil
	}

	return b.tg.Edit(b.userID, mid, validHTML, kb, tg.MarkupHTML)
}

// Replace last message + keyboard with the new ones
// Or show the new one (in case of image).
// Read "Markdown to HTML conversion" section in readme's ADRs
// Chat allows 1-4096 characters AFTER entities parsing,
// meaning we can have 4096 plain chars + any amount of tags.
func (b *Bot) showMD(probablyInvalidMD string, kb *tg.Keyboard) error {
	b.delAllImages()

	probablyInvalidMD, images, links := txt.ExtractTextImgsLinks(probablyInvalidMD)

	for label, link := range links {
		link = strings.TrimSpace(link)
		parts := strings.SplitN(link, "/", 2)
		dir := fs.DirUserRoot
		filename := link
		if len(parts) == 2 {
			dir = parts[0]
			filename = parts[1]
		}

		var cmd tg.Cmd
		mediaPath := link
		if len(parts) == 2 {
			mediaPath = parts[0] + "/" + parts[1]
		}
		if dir == fs.DirMedia || !strings.HasSuffix(strings.ToLower(mediaPath), fs.MDExt) {
			cmd = tg.NewCmd(CmdOpenMedia, []string{fs.ShortHash(mediaPath)})
		} else {
			cmd = tg.NewCmd(CmdShowFile, []string{fs.Hash(dir), fs.Hash(filename)})
		}
		kb.PrependRow(tg.NewRow(tg.NewBtn(txt.Ucfirst(label), cmd)))
	}

	mid, hasLastKeyboard := b.db.LastKeyboardMsgID()
	textChunks := txt.SplitTextIntoChunks(probablyInvalidMD, maxMsgLength)
	if !hasLastKeyboard || len(textChunks) > 1 || len(images) > 0 {
		b.delAllKeyboards()

		// Sending a gallery of images if there are any
		if len(images) > 0 {
			// We tolerate errors with the image gallery for now, text is more important
			mids, imgErr := b.tg.SendImages(b.userID, images)
			if imgErr == nil {
				for _, imgMid := range mids {
					b.db.AddImgMsgID(imgMid)
				}
			} else {
				slog.Error("Can't send images", "error", imgErr)
			}
		}

		// If our msg is too long, we send maxMsgsToSendAtOnce first messages.
		// Keyboard is attached to the last one
		textChunks = textChunks[0:min(maxMsgsToSendAtOnce, len(textChunks))]
		lastChunk := textChunks[len(textChunks)-1]
		textChunks = textChunks[0 : len(textChunks)-1]
		for _, textChunk := range textChunks {
			_, _ = b.tg.Send(b.userID, txt.MarkdownToHTML(textChunk), nil, tg.MarkupHTML)
		}

		mid, err := b.tg.Send(b.userID, txt.MarkdownToHTML(lastChunk), kb, tg.MarkupHTML)
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}

		b.db.SetLastKeyboardMsgID(mid)

		return nil
	}

	return b.tg.Edit(b.userID, mid, txt.MarkdownToHTML(probablyInvalidMD), kb, tg.MarkupHTML)
}

func (b *Bot) showMoveTo(params []string) error {
	msgHash := params[0]

	if b.cfg.NotesOnlyMode() {
		b.delAllKeyboards()

		return b.showMoveToFileOrDir([]string{msgHash})
	}

	var kb tg.Keyboard
	userMoveToBtns := b.moveToBtns(msgHash)
	if len(userMoveToBtns) == 0 {
		b.delAllKeyboards()

		return b.ShowHome(nil)
	}

	// Add recent command if any
	recentBtn := b.recentCmdBtn(msgHash)
	if recentBtn != nil {
		userMoveToBtns = append(userMoveToBtns, *recentBtn)
	}

	// This command is "do nothing and leave an item in the inbox"
	if !b.cfg.TasksOnlyMode() {
		showTodayCmd := tg.NewCmd(CmdShowHome, []string{})
		showTodayLabel := "👌"
		userMoveToBtns = append(userMoveToBtns, tg.NewBtn(showTodayLabel, showTodayCmd))
	}

	if !b.cfg.NotesOnlyMode() && !b.cfg.JournalOnlyMode() {
		prioBtns := b.priorityBtns(msgHash)
		if len(prioBtns) > 0 {
			kb.AddRow(tg.NewRow(prioBtns...))
		}
		_ = b.cfg.EnsureTaskCategories()
		catBtns := b.categoryBtns(msgHash)
		for _, row := range slice.Chunk(catBtns, btnsPerRow) {
			kb.AddRow(tg.NewRow(row...))
		}
	}

	userBtnsByRows := slice.Chunk(userMoveToBtns, btnsPerRow)
	for _, row := range userBtnsByRows {
		kb.AddRow(row)
	}

	b.delAllKeyboards()

	msg := b.tr("Saved!")
	if err := b.showHTML(msg, &kb); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	return nil
}

func (b *Bot) priorityBtns(msgHash string) []tg.Btn {
	var btns []tg.Btn
	for i, emoji := range b.cfg.PriorityEmojis() {
		btns = append(btns, tg.NewBtn(emoji, tg.NewCmd(CmdSetPriority, []string{msgHash, strconv.Itoa(i)})))
	}
	return btns
}

func (b *Bot) categoryBtns(msgHash string) []tg.Btn {
	var btns []tg.Btn
	for _, category := range b.cfg.TaskCategories() {
		filename := fs.SanitizeFilename(category) + "_.md"
		btns = append(btns, tg.NewBtn(category, tg.NewCmd(CmdMoveToChecklist, []string{fs.Hash(filename), msgHash})))
	}
	return btns
}

func (b *Bot) setPriority(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("set priority: missing params")
	}
	msgHash := params[0]
	idx, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("set priority: invalid index: %w", err)
	}

	emojis := b.cfg.PriorityEmojis()
	if idx < 0 || idx >= len(emojis) {
		return fmt.Errorf("set priority: index out of range")
	}

	chatMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("set priority: can't read chat: %w", err)
	}

	_, block, ok := findChatMsgByHash(chatMD, msgHash)
	if !ok {
		return fmt.Errorf("set priority: message not found")
	}

	newBody := priority.Apply(stripInboxEntryPrefix(block), emojis[idx], emojis)
	updated, err := renameChatMsg(chatMD, msgHash, newBody)
	if err != nil {
		return fmt.Errorf("set priority: %w", err)
	}
	if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, updated); err != nil {
		return fmt.Errorf("set priority: can't write chat: %w", err)
	}

	return b.showMoveTo([]string{msgHash})
}

func (b *Bot) showMorningSummary(_ []string) error {
	report, err := morningsummary.Build(b.fs, b.cfg)
	if err != nil {
		return fmt.Errorf("show morning summary: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))})
	return b.showHTML(report, kb)
}

func (b *Bot) recentCmdBtn(msgHash string) *tg.Btn {
	recentCmd, ok := b.db.RecentCommand()
	if !ok {
		return nil
	}

	args, _ := b.db.RecentCommandParams()
	args = append(args, msgHash)
	targetFilenameHash := args[0]

	var unhashedTarget string
	icon := "⭐️"
	if recentCmd == CmdMoveToExistingFile {
		var err error
		unhashedTarget, err = b.fs.Unhash(fs.DirUserRoot, targetFilenameHash)
		if err != nil {
			return nil
		}
	} else if recentCmd == CmdMoveToExistingNote {
		dir, err := b.fs.Unhash(fs.DirUserRoot, args[1])
		if err != nil {
			return nil
		}

		unhashedTarget, err = b.fs.Unhash(dir, targetFilenameHash)
		if err != nil {
			return nil
		}
	} else {
		return nil
	}

	name := fmt.Sprintf("%s %s", icon, fs.DisplayName(unhashedTarget))
	btn := tg.NewBtn(name, tg.NewCmd(recentCmd, args))
	return &btn
}

func (b *Bot) ShowHome(_ []string) error {
	report, err := morningsummary.Build(b.fs, b.cfg)
	if err != nil {
		report = ""
	}
	report = strings.TrimSpace(report)
	if report == "" {
		report = "🌴"
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr("📋 Tasks"), tg.NewCmd(CmdShowTasksView, nil)),
			tg.NewBtn(i18n.Tr("🗒 Notes"), tg.NewCmd(CmdShowNotesHub, nil)),
			tg.NewBtn(i18n.Tr("🌐 Spheres"), tg.NewCmd(CmdShowLifeSpheres, nil)),
		),
	})

	return b.showHTML(report, kb)
}

func (b *Bot) homeMessage(shownCount int) string {
	label := b.homeLabel(shownCount)

	report, err := morningsummary.Build(b.fs, b.cfg)
	if err != nil {
		return label
	}

	report = strings.TrimSpace(report)
	if report == "" {
		return label
	}

	return report + "\n\n" + label
}

func (b *Bot) showLaterTasks(_ []string) error {
	var kb tg.Keyboard

	// Adding tasks from Later.md
	laterChecklistMD, err := b.fs.Read(fs.DirUserRoot, fs.LaterFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("show later: can't read later file: %w", err)
	}
	if len(laterChecklistMD) != 0 {
		tasks := txt.IncompleteChecklistItems(laterChecklistMD)
		for _, task := range tasks {
			cmd := tg.NewCmd(CmdShowTask, []string{taskKindList, fs.Hash(fs.LaterFilename), fs.Hash(task)})
			btn := tg.NewBtn(taskPreviewLabel(task), cmd)
			kb.AddRow(btn)
		}
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	msg := b.tr("⏳ Your tasks for <b>later</b>:")
	err = b.showHTML(msg, &kb)
	if err != nil {
		return fmt.Errorf("show list: %w", err)
	}

	return nil
}

// TODO improve a bit
// msgsCount - how many messages (inbox items) were shown to a user
func (b *Bot) homeLabel(msgsCount ...int) string {
	var statusBar string

	hasPomodoroInToday := false
	todayMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		_, completed := txt.ChecklistItems(todayMD)
		checked, exists := completed[fs.PomodoroTask]
		hasPomodoroInToday = exists && !checked
	}
	if hasPomodoroInToday {
		statusBar = i18n.Emoji(fs.DisplayName(fs.PomodoroTask))
	}

	tasksCount := 0
	if len(msgsCount) > 0 && msgsCount[0] > 0 {
		tasksCount += msgsCount[0]
	}

	if tasksCount == 0 {
		statusBar += i18n.Emoji("palm")
	}

	if len(statusBar) != 0 {
		statusBar += " "
	}

	if tasksCount == 0 {
		return statusBar + i18n.Tr("Nothing here yet - send me something!")
	}

	postfix := i18n.Tr("items")
	if tasksCount == 1 {
		postfix = i18n.Tr("item")
	}

	return statusBar + fmt.Sprintf(i18n.Tr("<b>%d</b> %s%s"), tasksCount, postfix, wideSpacer)
}

func (b *Bot) randomNote(_ []string) error {
	rootEntries, err := b.fs.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return fmt.Errorf("random note: can't get root: %w", err)
	}
	type note struct {
		dir, name string
	}
	var notes []note
	for _, dir := range fs.OnlyNoteDirs(fs.OnlyDirs(rootEntries)) {
		entries, err := b.fs.FilesAndDirs(dir.Name)
		if err != nil {
			return fmt.Errorf("random note: can't get files in %s: %w", dir.Name, err)
		}
		for _, f := range fs.OnlyUserMDFiles(entries) {
			notes = append(notes, note{dir: dir.Name, name: f.Name})
		}
	}
	if len(notes) == 0 {
		return b.ShowHome(nil)
	}
	pick := notes[rand.Intn(len(notes))]
	return b.showFile([]string{fs.Hash(pick.dir), fs.Hash(pick.name)})
}

func (b *Bot) showFiles(_ []string) error {
	files, err := b.fs.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return fmt.Errorf("show files: can't get files: %w", err)
	}

	var kb tg.Keyboard
	mdFiles := fs.ExcludeConfig(fs.OnlyUserMDFiles(files))
	var fileBtns []tg.Btn
	for _, file := range mdFiles {
		cmd := tg.NewCmd(CmdShowFile, []string{fs.DirUserRoot, fs.Hash(file.Name)})
		btn := tg.NewBtn(fmt.Sprintf("%s", fs.UnsanitizeFilename(file.DisplayName)), cmd)
		fileBtns = append(fileBtns, btn)
	}
	fileBtnsByRows := slice.Chunk(fileBtns, btnsPerRow)
	for _, row := range fileBtnsByRows {
		kb.AddRow(row)
	}
	inlineCmd := tg.NewCustomCmd(CmdInlineQuerySearchEveryWhere, nil, tg.CmdTypeInlineQueryCurrentChat)

	footer := tg.NewRow(tg.NewBtn(i18n.Tr("🔎 Search"), inlineCmd))
	if !b.cfg.NotesOnlyMode() {
		footer = append(footer, tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))
	}
	kb.AddRow(footer)

	err = b.showHTML(b.tr("📄 Your files:")+wideSpacer, &kb)
	if err != nil {
		return fmt.Errorf("show files: %w", err)
	}

	return nil
}

func (b *Bot) showDirs(_ []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) showChecklists(_ []string) error {
	checklists, err := b.fs.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return fmt.Errorf("show checklists: %w", err)
	}
	checklists = fs.OnlyChecklists(checklists)

	var kb tg.Keyboard
	for _, checklist := range checklists {
		cmd := tg.NewCmd(CmdShowChecklist, []string{fs.Hash(checklist.Name)})
		btn := tg.NewBtn(i18n.AddEmoji(checklistTitle(checklist.Name)), cmd)

		kb.AddRow(btn)
	}
	kb.AddRow(tg.NewBtn(b.tr("🏠 Home"), tg.NewCmd(CmdShowHome, nil)))

	err = b.showHTML(b.tr("☑️ Checklists"), &kb)
	if err != nil {
		return fmt.Errorf("show checklists: %w", err)
	}

	return nil
}

func (b *Bot) showPostpone(_ []string) error {
	var kb tg.Keyboard

	// Inbox items also show in /postpone so the user can send them to Later.md.
	inboxMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		for _, block := range readChatMsgs(inboxMD) {
			if inboxHeaderRegex.MatchString(block) {
				continue
			}
			if strings.HasPrefix(block, "- [x] ") || strings.HasPrefix(block, "- [X] ") {
				continue
			}
			preview := strings.SplitN(stripInboxEntryPrefix(block), "\n", 2)[0]
			if len([]rune(preview)) > maxHeaderLengthForMobile {
				preview = string([]rune(preview)[:maxHeaderLengthForMobile]) + "…"
			}
			cmd := tg.NewCmd(CmdPostpone, []string{chatBlockHash(block)})
			kb.AddRow(tg.NewBtn("💬 "+preview, cmd))
		}
	}

	kb.AddRow(tg.NewRow(
		tg.NewBtn(b.tr("Rename"), tg.NewCmd(CmdShowRename, []string{})),
		tg.NewBtn(b.tr("OK"), tg.NewCmd(CmdShowHome, []string{})),
	))

	err = b.showHTML(b.tr("🦥 Select a task to postpone:"), &kb)
	if err != nil {
		return fmt.Errorf("show postpone: %w", err)
	}

	return nil
}

func (b *Bot) showMoveExisting(_ []string) error {
	var kb tg.Keyboard

	// Show today inbox items
	inboxContent, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		blocks := readChatMsgs(inboxContent)
		for _, block := range blocks {
			if inboxHeaderRegex.MatchString(block) {
				continue
			}
			// Skip already-completed entries — they're about to be swept anyway.
			if strings.HasPrefix(block, "- [x] ") || strings.HasPrefix(block, "- [X] ") {
				continue
			}
			preview := strings.SplitN(stripInboxEntryPrefix(block), "\n", 2)[0]
			if len([]rune(preview)) > maxHeaderLengthForMobile {
				preview = string([]rune(preview)[:maxHeaderLengthForMobile]) + "…"
			}
			cmd := tg.NewCmd(CmdShowMoveTo, []string{chatBlockHash(block)})
			kb.AddRow(tg.NewBtn("💬 "+preview, cmd))
		}
	}

	kb.AddRow(tg.NewRow(
		tg.NewBtn(b.tr("Rename"), tg.NewCmd(CmdShowRename, []string{})),
		tg.NewBtn(b.tr("OK"), tg.NewCmd(CmdShowHome, []string{})),
	))

	err = b.showHTML(b.tr("🦥 Select an item to move:"), &kb)
	if err != nil {
		return fmt.Errorf("show move from today: %w", err)
	}

	return nil
}

func (b *Bot) postpone(params []string) error {
	hash := params[0]

	err := b.moveFromChat(func(content string, _ time.Time) error {
		laterMD, rerr := b.fs.Read(fs.DirUserRoot, fs.LaterFilename)
		if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			return fmt.Errorf("postpone: can't read later file: %w", rerr)
		}
		return b.fs.Write(fs.DirUserRoot, fs.LaterFilename, txt.AddChecklistItem(laterMD, content, false))
	}, false, hash)
	if err != nil {
		return fmt.Errorf("postpone: can't move inbox entry to later: %w", err)
	}

	return b.showPostpone(nil)
}

func (b *Bot) showRename(_ []string) error {
	var kb tg.Keyboard

	inboxMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		for _, block := range readChatMsgs(inboxMD) {
			if inboxHeaderRegex.MatchString(block) {
				continue
			}
			if strings.HasPrefix(block, "- [x] ") || strings.HasPrefix(block, "- [X] ") {
				continue
			}
			preview := strings.SplitN(stripInboxEntryPrefix(block), "\n", 2)[0]
			if len([]rune(preview)) > maxHeaderLengthForMobile {
				preview = string([]rune(preview)[:maxHeaderLengthForMobile]) + "…"
			}
			cmd := tg.NewCmd(CmdShowRenameFile, []string{fs.ChatFilename, chatBlockHash(block)})
			kb.AddRow(tg.NewBtn("💬 "+preview, cmd))
		}
	}

	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	err = b.showHTML(b.homeLabel(), &kb)
	if err != nil {
		return fmt.Errorf("show rename: %w", err)
	}

	return nil
}

func (b *Bot) showRenameFile(params []string) error {
	checklist := params[0]
	itemHash := params[1]

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrBack), tg.NewCmd(CmdShowHome, []string{}))),
	})

	cmd := tg.NewCmd(CmdRename, []string{checklist, itemHash, "%s"})
	b.db.SetInputExpectation(cmd)

	err := b.showHTML(i18n.Tr("OK. Send me the new name for your task"), kb)
	if err != nil {
		return fmt.Errorf("show rename: %w", err)
	}

	return nil
}

func (b *Bot) rename(params []string) error {
	checklist := params[0]
	itemHash := params[1]
	newItemNameFromUserInput := params[2]

	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("rename: can't read checklist %s: %w", checklist, err)
	}

	if checklist == fs.ChatFilename {
		md, err = renameChatMsg(md, itemHash, newItemNameFromUserInput)
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	} else {
		md, _ = txt.RemoveChecklistItem(md, itemHash)
		md = txt.AddChecklistItem(md, newItemNameFromUserInput, false)
	}

	err = b.fs.Write(fs.DirUserRoot, checklist, md)
	if err != nil {
		return fmt.Errorf("rename: can't write checklist %s: %w", checklist, err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) showStats(_ []string) error {
	report, err := stats.TodayReport(b.fs, b.db, b.userID)
	if err != nil {
		return fmt.Errorf("show stats: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))})
	err = b.showHTML(strings.TrimSpace(report), kb)
	if err != nil {
		return fmt.Errorf("show stats: %w", err)
	}

	return nil
}

func (b *Bot) showSchedule(_ []string) error {
	scheduledTasks, err := b.cfg.Schedules()
	if err != nil {
		return fmt.Errorf("show schedule: %w", err)
	}
	schedule := ScheduleReport(scheduledTasks)
	if len(schedule) == 0 {
		schedule = i18n.Tr("You don't have any scheduled tasks! 🌴")
	}

	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))})
	err = b.showHTML(schedule, kb)
	if err != nil {
		return fmt.Errorf("show stats: %w", err)
	}

	return nil
}

func (b *Bot) showRead(_ []string) error {
	return b.showChecklist([]string{fs.Hash(fs.ReadFilename)})
}

func (b *Bot) showWatch(_ []string) error {
	return b.showChecklist([]string{fs.Hash(fs.WatchFilename)})
}

func (b *Bot) showShop(_ []string) error {
	return b.showChecklist([]string{fs.Hash(fs.ShopFilename)})
}

// TODO Chat.md move to today/later
func (b *Bot) showLongItemFromChecklist(params []string) error {
	checklistHash := params[0]
	itemHash := params[1]

	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("show checklist task: %w", err)
	}

	back := tg.NewCmd(CmdShowHome, nil)
	if checklist == fs.LaterFilename {
		back = tg.NewCmd(CmdShowTasksView, nil)
	} else if checklist != fs.ChatFilename {
		back = tg.NewCmd(CmdShowChecklist, []string{checklistHash})
	}
	return b.showListTaskWithBack(checklistHash, itemHash, back)
}

func (b *Bot) showLongItem(params []string) error {
	return b.showChatTask(params[0])
}

func (b *Bot) showFile(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("show file: can't find dir: %w", err)
	}

	// Inline-search results pass a nested path like
	// "habits/insights/2023 Habits.md" as the second param. Unhash only
	// resolves immediate children of a dir, so for any value containing
	// "/" we take it as a literal relative path.
	var filename string
	if strings.Contains(filenameHash, "/") {
		filename = filenameHash
	} else {
		filename, err = b.fs.Unhash(dir, filenameHash)
		if err != nil {
			return fmt.Errorf("show file: can't find file: %w", err)
		}
	}

	content, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("show file: %w", err)
	}

	isNotesDir := len(fs.OnlyNoteDirs([]fs.File{{Name: dir}})) > 0
	var kb *tg.Keyboard
	if life.IsDocDir(dir) || isNotesDir {
		kb = noteDetailKeyboard(dir, filename, dirHash)
	} else {
		row := tg.NewRow(tg.NewBtn("⬅️", noteBackCmd(dir)), tg.NewBtn("🏠", tg.NewCmd(CmdShowHome, nil)))
		kb = tg.NewKeyboard([]tg.Row{row})
	}

	if att, ok := txt.ParseAttachment(content); ok && txt.NeedsUserTitle(content) {
		return b.showAttachmentFile(att, kb)
	}

	displayContent := txt.FormatNoteDetailBody(content)

	md := fmt.Sprintf("**%s**\n\n%s", fs.DisplayName(filename), displayContent)
	err = b.showMD(md, kb)
	if err != nil {
		return fmt.Errorf("show file: %w", err)
	}

	msgID, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if hasLastKeyboard {
		b.db.SetHashOrPathByMsgID(msgID, filepath.ToSlash(filepath.Join(dir, filename)))
	}

	return nil
}

func (b *Bot) showAttachmentFile(att txt.AttachmentInfo, kb *tg.Keyboard) error {
	return b.sendMediaDocument(att, kb)
}

func (b *Bot) openMedia(params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("open media: missing params")
	}
	mediaPath, err := b.resolveMediaPath(params[0])
	if err != nil {
		return fmt.Errorf("open media: %w", err)
	}
	att := txt.AttachmentInfo{Path: mediaPath}
	return b.sendMediaDocument(att, nil)
}

func (b *Bot) resolveMediaPath(hashOrPath string) (string, error) {
	hashOrPath = strings.TrimSpace(hashOrPath)
	if strings.Contains(hashOrPath, "/") {
		return strings.TrimPrefix(hashOrPath, "/"), nil
	}
	files, err := b.fs.FilesAndDirs(fs.DirMedia)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.IsDir {
			continue
		}
		full := fs.DirMedia + "/" + f.Name
		if fs.ShortHash(full) == hashOrPath || fs.ShortHash(f.Name) == hashOrPath {
			return full, nil
		}
	}
	return "", fmt.Errorf("media not found: %s", hashOrPath)
}

func (b *Bot) sendMediaDocument(att txt.AttachmentInfo, kb *tg.Keyboard) error {
	displayName := txt.AttachmentDisplayName(att.Name, att.Path)
	dir, filename := txt.AttachmentMediaPath(att.Path)
	data, err := b.fs.Read(dir, filename)
	if err != nil {
		md := fmt.Sprintf("**%s**\n\n📎 %s", displayName, att.Path)
		return b.showMD(md, kb)
	}

	b.delAllKeyboards()
	mid, err := b.tg.SendDocument(b.userID, displayName, strings.NewReader(data), displayName, kb)
	if err != nil {
		return fmt.Errorf("show attachment: %w", err)
	}

	b.db.SetLastKeyboardMsgID(mid)
	msgID, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if hasLastKeyboard {
		b.db.SetHashOrPathByMsgID(msgID, att.Path)
	}
	return nil
}

func (b *Bot) showChecklist(params []string) error {
	checklistHash := params[0]

	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("show checklist: %w", err)
	}

	md, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("show checklist: %w", err)
	}

	items := txt.IncompleteChecklistItems(md)
	maxButtons := maxBtns
	if checklist == fs.ReadFilename || checklist == fs.WatchFilename {
		maxButtons = maxBtnsInChecklist
	}
	if len(items) > maxButtons {
		items = items[:maxButtons]
	}

	kb := tg.NewKeyboard(nil)
	for _, item := range items {
		if len([]rune(item)) >= maxHeaderLengthForMobile {
			cmd := tg.NewCmd(CmdShowLongItemFromChecklist, []string{fs.Hash(checklist), fs.Hash(item)})
			btn := tg.NewBtn(txt.Emoji(i18n.Emoji("eyes"), item), cmd)
			kb.AddRow(btn)
		} else {
			cmd := tg.NewCmd(CmdCompleteChecklistItem, []string{checklistHash, fs.Hash(item)})
			btn := tg.NewBtn(taskBtnLabel(item), cmd)
			kb.AddRow(btn)
		}
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	title := checklistTitle(checklist)
	err = b.showHTML(title+wideSpacer, kb)
	if err != nil {
		return fmt.Errorf("show checklist: %w", err)
	}

	return nil
}

func (b *Bot) showStart(params []string) error {
	// For now it is not used
	if len(params) > 0 {
		mode := strings.ToLower(params[0])
		if mode == "notes" {
			return b.setNotesOnlyMode(nil)
		} else if mode == "tasks" {
			return b.setTasksOnlyMode(nil)
		} else if mode == "journal" {
			return b.setJournalOnlyMode(nil)
		} else if mode == "full" {
			return b.setFullMode(nil)
		}
	}

	// We can tolerate an error.
	_ = b.setFullMode(nil)

	_, err := b.tg.Send(b.userID, i18n.Tr("Welcome! 👋\n\n<b>Присылай что угодно — выбери задачу или заметку.</b>"), nil, tg.MarkupHTML)

	return err
}

func (b *Bot) moveToDir(params []string) error {
	// TODO Remove input expectations if dir is not today
	toDirHash := params[0]

	msgHashes := strings.Split(params[1], ",")

	toDir, err := b.fs.FindNoteDirByShortHash(toDirHash)
	canCreateMissingDir := slices.Contains([]string{
		fs.DirArchive, fs.DirHabits, life.DirSpheres,
	}, toDirHash)
	if err != nil {
		if canCreateMissingDir {
			toDir = toDirHash
		} else {
			return fmt.Errorf("move: can't resolve dir %s: %w", toDirHash, err)
		}
	}

	err = b.moveFromChat(func(content string, timestamp time.Time) error {
		var sanitizedTitle string
		sanitizedTitle, content, err = b.extractHeaderAndBodyPreserveMedia(content, maxHeaderLength)
		if err != nil {
			return fmt.Errorf("move to dir from chat: can't extract title and content: %w", err)
		}

		filename := fs.Filename(sanitizedTitle)

		notesDir := fs.OnlyNoteDirs([]fs.File{{Name: toDir}})
		isNotesDir := len(notesDir) == 1
		if isNotesDir {
			// We can tolerate this, as this is informative logging
			_ = journal.AddRecord(b.fs, fmt.Sprintf("📌 %s", fs.DisplayName(filename)), b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())
		}

		return b.createOrAdd(toDir, filename, content)
	}, true, msgHashes...)

	b.delAllKeyboards()
	msg := txt.Emoji(i18n.Emoji("dir"), fmt.Sprintf(i18n.Tr("Moved to <b>%s</b>"), fs.DisplayName(toDir)))
	// Just an informative messages
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) moveToChecklist(params []string) error {
	toChecklistHash := params[0]

	msgHashes := strings.Split(params[1], ",")

	for _, msgHash := range msgHashes {
		_, err := b.addToChecklist(toChecklistHash, msgHash)
		if err != nil {
			return fmt.Errorf("move to checklist: can't add to checklist: %w", err)
		}
	}

	return b.ShowHome(nil)
}

func (b *Bot) addToChecklist(checklistHash string, msgHash string) (string, error) {
	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	// Create known checklist if it doesn't exist
	if err != nil {
		supportedChecklists := []string{
			fs.ChatFilename,
			fs.LaterFilename,
			fs.ReadFilename,
			fs.WatchFilename,
			fs.ShopFilename,
		}

		created := false
		for _, supportedChecklist := range supportedChecklists {
			if fs.Hash(supportedChecklist) == checklistHash || supportedChecklist == checklistHash {
				checklist = supportedChecklist
				err = b.fs.Write(fs.DirUserRoot, checklist, "")
				if err != nil {
					return "", fmt.Errorf("add to checklist: can't create checklist %s: %w", checklist, err)
				}
				created = true
				break
			}
		}

		if !created {
			return "", fmt.Errorf("add to checklist: can't unhash checklist %s: %w", checklistHash, err)
		}
	}

	checklistMD, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return "", fmt.Errorf("add to checklist: can't read checklist %s: %w", checklist, err)
	}

	var item string
	err = b.moveFromChat(func(content string, timestamp time.Time) error {
		item = content
		md := txt.AddChecklistItem(checklistMD, content, false)
		return b.fs.Write(fs.DirUserRoot, checklist, md)
	}, true, msgHash)
	if err != nil {
		return "", fmt.Errorf("move to checklist: can't move from chat: %w", err)
	}

	return item, nil
}

func (b *Bot) completeChecklistItem(params []string) error {
	checklistHash := params[0]
	itemHash := params[1]

	checklist, err := b.fs.Unhash(fs.DirUserRoot, checklistHash)
	if err != nil {
		return fmt.Errorf("complete checklist item: can't unhash checklist %s: %w", checklistHash, err)
	}

	checklistMD, err := b.fs.Read(fs.DirUserRoot, checklist)
	if err != nil {
		return fmt.Errorf("complete checklist item: can't read checklist %s: %w", checklist, err)
	}

	md, item := txt.CompleteChecklistItem(checklistMD, itemHash)
	err = b.fs.Write(fs.DirUserRoot, checklist, md)
	if err != nil {
		return fmt.Errorf("complete checklist item: can't complete item from chat: %w", err)
	}

	// We can tolerate failure of writing to journal, since that's not single source of truth.
	// AddRecord prepends a fresh `HH:MM`; strip any leading timestamp on
	// the item body so we don't end up with two of them.
	_ = journal.AddRecord(b.fs, fmt.Sprintf("✅ %s", fs.DisplayName(txt.StripChatTimestamp(item))), b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())

	if checklist == fs.LaterFilename {
		return b.showTasksView(nil)
	} else if checklist != fs.ChatFilename {
		return b.showChecklist([]string{checklist})
	}

	return b.ShowHome(nil)
}

func (b *Bot) requestNewDirName(params []string) error {
	filenameHash := params[0]

	err := b.showHTML(i18n.Tr("OK. Send me the name for your new dir"), nil)
	if err != nil {
		return fmt.Errorf("request new dir: %w", err)
	}

	b.db.SetInputExpectation(tg.NewCmd(CmdMoveToNewDir, []string{filenameHash, "%s"}))

	return nil
}

// moveToNewDir accepts dir name as a second parameter
// which is a bit off, but the thing is sometimes it is replaced with
// inputExpectation, which only can add parameters in the end.
func (b *Bot) moveToNewDir(params []string) error {
	msgIndicesStr := params[0]
	dir := sanitizeDirPath(params[1])

	exists, err := b.fs.Exists(dir, "")
	if err != nil {
		return fmt.Errorf("move to new dir from caht: %w", err)
	}
	if !exists {
		err = b.fs.MakeDir(dir)
		if err != nil {
			return fmt.Errorf("move to new dir from chat: %w", err)
		}
	}

	return b.moveToDir([]string{dir, msgIndicesStr})
}

func sanitizeDirPath(dir string) string {
	dir = strings.TrimSpace(strings.ReplaceAll(dir, "\\", "/"))
	parts := strings.Split(dir, "/")
	for i, part := range parts {
		parts[i] = strings.ToLower(fs.SanitizeFilename(part))
	}
	return strings.Join(parts, "/")
}

// TODO reuse move to existing note as more general?
func (b *Bot) moveToExistingFile(params []string) error {
	// TODO Remove input expectations if dir is not today (?)
	existingFilenameHash := params[0]

	msgHashes := strings.Split(params[1], ",")

	existingFilename, err := b.fs.Unhash(fs.DirUserRoot, existingFilenameHash)
	if err != nil {
		return fmt.Errorf("move to file: can't unhash existing file '%s': %w", existingFilenameHash, err)
	}

	err = b.moveFromChat(func(content string, timestamp time.Time) error {
		return b.addToFile(fs.DirUserRoot, existingFilename, content)
	}, true, msgHashes...)
	if err != nil {
		return fmt.Errorf("move to file: can't add to existing file '%s': %w", existingFilename, err)
	}

	b.db.SetRecentCommand(CmdMoveToExistingFile)
	b.db.SetRecentCommandParams([]string{fs.ShortHash(existingFilename)})

	b.delAllKeyboards()
	msg := txt.Emoji(i18n.Emoji("file"), fmt.Sprintf(i18n.Tr("Saved to <b>%s</b>"), fs.DisplayName(existingFilename)))
	// Just an informative messages
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) moveToExistingNote(params []string) error {
	toFilenameHash := params[0]
	toDirHash := params[1]

	msgHashes := strings.Split(params[2], ",")

	var toDir string
	if toDirHash == "" {
		toDir = fs.DirUserRoot
	} else {
		var err error
		toDir, err = b.fs.FindNoteDirByShortHash(toDirHash)
		if err != nil {
			return fmt.Errorf("move to existing note: %w", err)
		}
	}

	toFilename, err := b.fs.Unhash(toDir, toFilenameHash)
	if err != nil {
		return fmt.Errorf("move to existing note:: %w", err)
	}

	err = b.moveFromChat(func(content string, t time.Time) error {
		err = b.addToFile(toDir, toFilename, content)
		if err != nil {
			return fmt.Errorf("move to existing note: can't add to file %s: %w", toFilename, err)
		}

		b.db.SetRecentCommand(CmdMoveToExistingNote)
		b.db.SetRecentCommandParams([]string{fs.ShortHash(toFilename), fs.ShortHash(toDir)})

		return nil
	}, false, msgHashes...)
	if err != nil {
		return fmt.Errorf("move to existing note: can't read content from chat: %w", err)
	}

	b.delAllKeyboards()
	msg := txt.Emoji(i18n.Emoji("file"), fmt.Sprintf(i18n.Tr("Saved to <b>%s</b>"), fs.DisplayName(toFilename)))
	// Just an informative messages
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) moveToDirChecklist(params []string) error {
	msgHashes := strings.Split(params[0], ",")
	checklistDirHash := params[1]

	checklistDir, err := b.fs.Unhash(fs.DirUserRoot, checklistDirHash)
	if err != nil {
		return fmt.Errorf("move to checklistDir: %w", err)
	}

	err = b.moveFromChat(func(content string, t time.Time) error {
		isMultiline := txt.IsMultiline(content)

		if isMultiline && b.cfg.ShouldSplitChecklist(checklistDir) {
			content = strings.TrimSpace(txt.NormNewLines(content))
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = fs.SanitizeFilename(line)
				err = b.fs.Write(checklistDir, fs.Filename(line), "")
				if err != nil {
					return fmt.Errorf("move to checklistDir: %w", err)
				}
			}
		} else {
			sanitizedTitle, content, err := b.extractHeaderAndBody(content, maxHeaderLengthForMobile)
			if err != nil {
				return fmt.Errorf("move to checklistDir: %w", err)
			}
			filename := fs.Filename(sanitizedTitle)
			return b.fs.Write(checklistDir, filename, content)
		}

		return nil
	}, false, msgHashes...)
	if err != nil {
		return fmt.Errorf("move to checklistDir: can't read content from chat: %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) moveToRead(params []string) error {
	msgIndices := params[0]

	return b.moveToChecklist([]string{fs.Hash(fs.ReadFilename), msgIndices})
}

func (b *Bot) moveToWatch(params []string) error {
	msgIndices := params[0]

	return b.moveToChecklist([]string{fs.Hash(fs.WatchFilename), msgIndices})
}

func (b *Bot) moveToShop(params []string) error {
	msgIndices := params[0]

	return b.moveToChecklist([]string{fs.Hash(fs.ShopFilename), msgIndices})
}

func (b *Bot) moveToNewFile(params []string) error {
	msgHash := params[0]
	newFilenameFromUserInput := fs.Filename(params[1])

	//filename, err := b.fs.Unhash(fs.DirUserRoot, msgIndex)
	//if err != nil {
	//	return fmt.Errorf("move to new file: can't unhash existing file '%s': %w", msgIndex, err)
	//}
	//
	//// Save existing filename to content in case the content of new file is empty (i.e. not multiline)
	//content, err := b.fs.Read(fs.DirUserRoot, filename)
	//if err != nil {
	//	return fmt.Errorf("move to new file: can't read file '%s': %w", filename, err)
	//}
	err := b.moveFromChat(func(content string, t time.Time) error {
		content = strings.TrimSpace(content)
		//if len(content) == 0 {
		//	content = fs.DisplayName(filename)
		//	err = b.fs.Write(fs.DirUserRoot, filename, content)
		//	if err != nil {
		//		return fmt.Errorf("move to new file: can't write content of '%s': %w", filename, err)
		//	}
		//}

		// TODO check for safety
		// TODO won't we lost some text here in case of multiline?
		//err = b.fs.Rename(fs.DirUserRoot, filename, fs.DirUserRoot, newFilenameFromUserInput)
		//if err != nil {
		//	return fmt.Errorf("move to new file: can't create empty file: %w", err)
		//}

		// We can tolerate this
		//_ = journal.AddRecord(b.fs, fmt.Sprintf("📄 %s", fs.DisplayName(filename)), b.cfg.Timezone())

		b.db.SetRecentCommand(CmdMoveToExistingFile)
		b.db.SetRecentCommandParams([]string{fs.ShortHash(newFilenameFromUserInput)})

		// TODO add if exists
		return b.fs.Write(fs.DirUserRoot, newFilenameFromUserInput, content)
	}, false, msgHash)
	if err != nil {
		return fmt.Errorf("move to new file: can't read content from chat: %w", err)
	}

	msg := txt.Emoji(i18n.Emoji("file"), fmt.Sprintf(i18n.Tr("Saved to <b>%s</b>"), fs.DisplayName(newFilenameFromUserInput)))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) moveToNewChecklist(params []string) error {
	msgHash := params[0]

	supposedName := params[1]
	supposedName = fs.SanitizeFilename(supposedName)

	dir := strings.ToLower(supposedName)
	dir = fmt.Sprintf("_%s_", dir)
	exists, err := b.fs.Exists(fs.DirUserRoot, dir)
	if err != nil {
		return fmt.Errorf("move to new checklist: %w", err)
	}
	if !exists {
		err = b.fs.MakeDir(dir)
	}

	return b.moveToDir([]string{dir, msgHash})
}

func (b *Bot) moveToJournal(params []string) error {
	msgHashes := params

	err := b.moveFromChat(func(content string, t time.Time) error {
		// TODO take into account time from chat
		return journal.AddRecord(b.fs, content, b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())
	}, false, msgHashes...)
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't add record: %w", err)
	}

	b.delAllKeyboards()
	msg := txt.Emoji(i18n.Emoji("journal"), i18n.Tr("Saved to <b>journal</b>"))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	if b.cfg.JournalOnlyMode() {
		return nil
	}

	return b.ShowHome(nil)
}

func (b *Bot) addToJournalAndContinue(params []string) error {
	content := params[0]

	err := journal.AddRecord(b.fs, content, b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't add note: %w", err)
	}

	// Don't return - continue to save to inbox as well.
	msgHash, err := b.appendToChat(content, b.cfg.Timezone())
	if err != nil {
		return fmt.Errorf("save to inbox: %w", err)
	}

	return b.showMoveTo([]string{msgHash})
}

func (b *Bot) addToJournalFromShortcut(params []string) error {
	content := params[0]

	// TODO change to pass text
	err := journal.AddRecord(b.fs, content, b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't add note: %w", err)
	}

	msg := i18n.Tr("Saved to <b>Journal</b>")
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

// TODO add tests
func (b *Bot) addToRecentFileOrNoteFromShortcut(params []string) error {
	content := params[0]

	args, _ := b.db.RecentCommandParams()
	cmd, _ := b.db.RecentCommand()

	var existingFilename string
	if cmd == CmdMoveToExistingFile {
		var err error
		existingFilename, err = b.fs.Unhash(fs.DirUserRoot, args[0])
		if err != nil {
			return fmt.Errorf("failed to move to recent file or note: can't unhash filename: %w", err)
		}

		err = b.addToFile(fs.DirUserRoot, existingFilename, content)
		if err != nil {
			return fmt.Errorf("failed to move to recent file: can't add note: %w", err)
		}
	} else if cmd == CmdMoveToExistingNote {
		dir, err := b.fs.Unhash(fs.DirUserRoot, args[1])
		if err != nil {
			return fmt.Errorf("failed to move to recent note: can't unhash dir: %w", err)
		}
		existingFilename, err = b.fs.Unhash(dir, args[0])
		if err != nil {
			return fmt.Errorf("failed to move to recent note: can't unhash filename: %w", err)
		}

		err = b.addToFile(dir, existingFilename, content)
		if err != nil {
			return fmt.Errorf("failed to move to recent note: can't add note: %w", err)
		}
	} else {
		return nil
	}

	msg := fmt.Sprintf(i18n.Tr("Added to <b>%s</b>"), fs.DisplayName(existingFilename))
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) moveToLater(params []string) error {
	msgHash := params[0]

	return b.moveToChecklist([]string{fs.LaterFilename, msgHash})
}

// complete marks a Chat.md entry as done (`[ ]` → `[x]`), keeping it in the file.
func (b *Bot) complete(params []string) error {
	msgHash := params[0]

	key, err := b.fs.SafePath(fs.DirUserRoot, "")
	if err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	lock := userLock(key)
	lock.Lock()
	defer lock.Unlock()

	content, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return fmt.Errorf("complete: can't read inbox: %w", err)
	}

	newContent, item, flipped, err := completeChatMsg(content, msgHash)
	if err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	if !flipped {
		return b.showTasksView(nil)
	}

	if err := b.fs.Write(fs.DirUserRoot, fs.ChatFilename, newContent); err != nil {
		return fmt.Errorf("complete: can't write inbox: %w", err)
	}

	_ = journal.AddRecord(b.fs, fmt.Sprintf("✅ %s", fs.DisplayName(txt.StripChatTimestamp(item))), b.cfg.Timezone(), b.cfg.JournalTimestampsEnabled())

	return b.showTasksView(nil)
}

func (b *Bot) completeListItem(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.Unhash(fs.DirUserRoot, dirHash)
	if err != nil {
		return fmt.Errorf("complete: can't unhash dir %s: %w", dir, err)
	}

	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("complete: can't unhash filename %s: %w", filename, err)
	}

	if err = b.fs.Touch(dir, filename); err != nil {
		return fmt.Errorf("complete: can't touch %s: %w", filename, err)
	}

	err = b.fs.Rename(dir, filename, fs.DirArchive, filename)
	if err != nil {
		return fmt.Errorf("complete: can't complete %s: %w", filename, err)
	}

	return b.showChecklist([]string{dirHash})
}

func (b *Bot) showChecklistItem(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.Unhash(fs.DirUserRoot, dirHash)
	if err != nil {
		return fmt.Errorf("show checklist item: can't unhash dir %s: %w", dir, err)
	}

	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("show checklist item: can't unhash filename %s: %w", filename, err)
	}

	content, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("show checklist item: can't read content of %s: %w", filename, err)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn(i18n.Tr(i18n.StrBack), tg.NewCmd(CmdShowChecklist, []string{dirHash})),
			tg.NewBtn(i18n.Tr(i18n.StrComplete), tg.NewCmd(CmdCompleteListItem, []string{dirHash, filenameHash})),
		),
	})

	err = b.showHTML(content, kb)
	if err != nil {
		return fmt.Errorf("show checklist item: %w", err)
	}

	msgID, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if hasLastKeyboard {
		b.db.SetHashOrPathByMsgID(msgID, filepath.ToSlash(filepath.Join(dir, filename)))
	}

	return nil
}

func (b *Bot) schedule(params []string) error {
	msgHash := params[0]
	timeStr := params[1]
	cron := params[2]

	scheduleTime, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("schedule: can't parse timestamp: %w", err)
	}

	item, err := b.addToChecklist(fs.LaterFilename, msgHash)
	if err != nil {
		return fmt.Errorf("schedule: can't move to later: %w", err)
	}

	err = b.cfg.AddToSchedule(item, scheduleTime, cron)
	if err != nil {
		return fmt.Errorf("schedule: can't add to schedule: %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) scheduleForTmrw(params []string) error {
	return b.schedule([]string{params[0], txt.I64(Tomorrow()), ""})
}

func (b *Bot) delAllKeyboards() {
	var msgIDs []int
	mid, hasLastKeyboard := b.db.LastKeyboardMsgID()
	if hasLastKeyboard {
		b.db.DelLastKeyboardMsgID()
		msgIDs = append(msgIDs, mid)
	}

	// No worries if we fail - it will be cleaned up by the worker
	for _, msgID := range msgIDs {
		// If we fail to del - user would get a bunch
		// of keyboards in one chat, which is messy but not critical
		_ = b.tg.Del(b.userID, msgID)
	}
}

func (b *Bot) delAllImages() {
	mids, hasSentImages := b.db.ImgMsgID()
	if !hasSentImages {
		return
	}

	b.db.DelImgMsgID()
	for _, mid := range mids {
		// If we fail to del - user would get a bunch
		// of keyboards in one chat, which is messy but not critical
		_ = b.tg.Del(b.userID, mid)
	}
}

func (b *Bot) showToADay(params []string) error {
	filenameHash := params[0]

	kb, err := b.toADayKeyboard(filenameHash)
	if err != nil {
		return fmt.Errorf("show for a day: %w", err)
	}

	err = b.showHTML(i18n.Tr("Choose a day"), kb)
	if err != nil {
		return fmt.Errorf("show for a day: %w", err)
	}

	return nil
}

func (b *Bot) toADayKeyboard(filenameHash string) (*tg.Keyboard, error) {
	newBtn := func(name, cron string) tg.Btn {
		return tg.NewBtn(name, tg.NewCmd(CmdSchedule, []string{filenameHash, txt.I64(NextExcludeToday(cron)), ""}))
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewRow(tg.NewBtn(i18n.Tr(i18n.StrRepeat), tg.NewCmd(CmdShowScheduleForDayRecurring, []string{filenameHash}))),
		tg.NewRow(
			newBtn(i18n.Tr(i18n.StrMonday), "0 0 * * 1"),
			newBtn(i18n.Tr(i18n.StrTuesday), "0 0 * * 2"),
			newBtn(i18n.Tr(i18n.StrWednesday), "0 0 * * 3"),
			newBtn(i18n.Tr(i18n.StrThursday), "0 0 * * 4"),
		),
		tg.NewRow(
			newBtn(i18n.Tr(i18n.StrFriday), "0 0 * * 5"),
			newBtn(i18n.Tr(i18n.StrSaturday), "0 0 * * 6"),
			newBtn(i18n.Tr(i18n.StrSunday), "0 0 * * 0"),
		),
	})

	for _, iAndj := range [][]int{{1, 8}, {9, 16}, {17, 24}, {25, 31}} {
		row := tg.NewRow()
		for i := iAndj[0]; i <= iAndj[1]; i++ {
			cron := fmt.Sprintf("0 0 %d * *", i)
			row = append(row, newBtn(txt.I64(int64(i)), cron))
		}
		kb.AddRow(row)
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrToToday), tg.NewCmd(CmdShowHome, nil)))

	return kb, nil
}

func (b *Bot) showMoveToFileOrDir(params []string) error {
	msgHash := params[0]
	maxRecentBtns := maxGroupedBtnsInMoveTo

	userWantedAllBtns := len(params) > 1
	if userWantedAllBtns {
		maxRecentBtns = maxBtns
	} else {
		//b.db.SetRecentCommand(CmdMoveToExistingFile)
		//b.db.SetRecentCommandParams([]string{fs.ShortHash(filename), fs.ShortHash(fs.DirToday)})
	}

	kb := tg.NewKeyboard(nil)
	skippedBtns := false

	//fileBtns, err := b.moveToFileBtns(fs.ShortHash(filename))
	fileBtns, err := b.moveToFileBtns(msgHash)
	if err != nil {
		return fmt.Errorf("to file dialog: %w", err)
	}
	if len(fileBtns) > maxRecentBtns {
		fileBtns = fileBtns[:maxRecentBtns]
		skippedBtns = true
	}
	// Move newly created file to the end of the files list
	if len(fileBtns) > 0 {
		fileBtns = append(fileBtns[1:], fileBtns[0])
	}
	fileBtnsByRows := slice.Chunk(fileBtns, btnsPerRow)
	for _, row := range fileBtnsByRows {
		kb.AddRow(row)
	}

	searchCMD := tg.NewCustomCmd(CmdInlineQuerySearchEveryWhere, nil, tg.CmdTypeInlineQueryCurrentChat)
	kb.AddRow(tg.NewBtn(i18n.Tr("Search"), searchCMD))

	if skippedBtns {
		kb.AddRow(tg.NewBtn(i18n.Tr("More..."), tg.NewCmd(CmdShowMoveToDirOrFile, []string{msgHash, "full"})))
	}

	b.db.SetInputExpectation(tg.NewCmd(CmdMoveToNewFile, []string{msgHash, "%s"}))

	err = b.showHTML(i18n.Tr("📄 Select where to save or send a new name:"), kb)
	if err != nil {
		return fmt.Errorf("to file dialog: %w", err)
	}

	return nil
}

func (b *Bot) showToChecklist(params []string) error {
	filenameHash := params[0]

	kb, err := b.toChecklistKeyboard(filenameHash)
	if err != nil {
		return fmt.Errorf("show to checklist: can't get keyboard: %w", err)
	}

	b.db.SetInputExpectation(tg.NewCmd(CmdMoveToNewChecklist, []string{filenameHash, "%s"}))

	err = b.showHTML(i18n.Tr("Choose a checklist or name a new one"), kb)
	if err != nil {
		return fmt.Errorf("show to checklist: %w", err)
	}

	return nil
}

func (b *Bot) moveToFileBtns(msgHash string) ([]tg.Btn, error) {
	files, err := b.fs.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return nil, fmt.Errorf("to doc keyboard: %w", err)
	}
	files = fs.OnlyUserMDFiles(files)
	files = fs.SortByCtimeDesc(files)
	if len(files) == 0 {
		return nil, nil
	}

	var buttons []tg.Btn
	newBtn := func(title, existingFilenameHash string) tg.Btn {
		title = fmt.Sprintf("%s", title)
		params := []string{existingFilenameHash, msgHash}
		return tg.NewBtn(title, tg.NewCmd(CmdMoveToExistingFile, params))
	}
	for _, file := range files {
		buttons = append(buttons, newBtn(file.DisplayName, fs.ShortHash(file.Name)))
	}

	return buttons, nil
}

func (b *Bot) moveToDirBtns(msgHash string) ([]tg.Btn, error) {
	newBtn := func(dir string) tg.Btn {
		emojifiedDir := fmt.Sprintf("%s %s", i18n.Emoji("dir"), txt.Ucfirst(dir))
		return tg.NewBtn(emojifiedDir, tg.NewCmd(CmdMoveToExistingDir, []string{fs.ShortHash(dir), msgHash}))
	}

	dirPaths, err := b.fs.AllNoteDirPaths()
	if err != nil {
		return nil, fmt.Errorf("To File keyboard: %w", err)
	}

	var buttons []tg.Btn
	for _, dirPath := range dirPaths {
		buttons = append(buttons, newBtn(dirPath))
	}

	return buttons, nil
}

func (b *Bot) toChecklistKeyboard(filenameHash string) (*tg.Keyboard, error) {
	newBtn := func(dir, title string) tg.Btn {
		return tg.NewBtn(title, tg.NewCmd(CmdMoveToDirChecklist, []string{filenameHash, dir}))
	}

	dirs, err := b.fs.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return nil, fmt.Errorf("to checklist keyboard: %w", err)
	}
	// TODO handle case with zero dirs (inline_keyboard is null), for all similar cases
	dirs = fs.OnlyChecklists(fs.OnlyDirs(dirs))

	kb := tg.NewKeyboard(nil)
	for _, dir := range dirs {
		kb.AddRow(newBtn(dir.Name, dir.DisplayName))
	}

	return kb, nil
}

func (b *Bot) togglePomodoro(_ []string) error {
	// Check if Pomodoro is already running
	hasPomodoroInToday := false
	todayMD, err := b.fs.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		_, isCompleted := txt.ChecklistItems(todayMD)
		_, hasPomodoroInToday = isCompleted[fs.PomodoroTask]
	}

	if hasPomodoroInToday {
		todayMD, _ = txt.RemoveChecklistItem(todayMD, fs.PomodoroTask)
		err = b.fs.Write(fs.DirUserRoot, fs.ChatFilename, todayMD)
		if err != nil {
			return fmt.Errorf("toggle pomodoro: failed to delete pomodoro file: %w", err)
		}
	}

	if hasPomodoroInToday {
		_, _ = b.tg.Send(b.userID, i18n.Tr("Pomodoro is stopped"), nil, tg.MarkupHTML)
		return b.ShowHome(nil)
	}

	// Create Pomodoro checklist item
	err = b.fs.Write(fs.DirUserRoot, fs.ChatFilename, txt.AddChecklistItem(todayMD, fs.PomodoroTask, false))

	_, err = b.tg.Send(b.userID, i18n.Tr(i18n.PomodoroStarted), nil, tg.MarkupHTML)
	if err != nil {
		return fmt.Errorf("toggle pomodoro: failed to show pomodoro hint message %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) showToADayRecurring(params []string) error {
	filenameHash := params[0]

	newBtn := func(name, cron string) tg.Btn {
		// We need to shorten filehash, otherwise whole payload doesn't fit telegram's restrictions (64 bytes)
		cmd := tg.NewCmd(CmdSchedule, []string{txt.Substr(filenameHash, 0, 4), txt.I64(NextExcludeToday(cron)), cron})
		return tg.NewBtn(name, cmd)
	}

	kb := tg.NewKeyboard([]tg.Row{
		// Cron format: Minute Hour DayOfMonth Month DayOfWeek
		tg.NewRow(
			newBtn(i18n.Tr(i18n.StrWeekdays), "0 0 * * 1-5"),
			newBtn(i18n.Tr(i18n.StrEveryday), "0 0 * * *"),
		),
		tg.NewRow(
			newBtn(i18n.Tr(i18n.StrMonday), "0 0 * * 1"),
			newBtn(i18n.Tr(i18n.StrTuesday), "0 0 * * 2"),
			newBtn(i18n.Tr(i18n.StrWednesday), "0 0 * * 3"),
			newBtn(i18n.Tr(i18n.StrThursday), "0 0 * * 4"),
		),
		tg.NewRow(
			newBtn(i18n.Tr(i18n.StrFriday), "0 0 * * 5"),
			newBtn(i18n.Tr(i18n.StrSaturday), "0 0 * * 6"),
			newBtn(i18n.Tr(i18n.StrSunday), "0 0 * * 0"),
		),
	})

	for week := 0; week < 4; week++ {
		row := tg.NewRow()
		for day := 1; day < 8; day++ {
			i := week*7 + day
			cron := fmt.Sprintf("0 0 %d * *", i)
			row = append(row, newBtn(txt.I64(int64(i)), cron))
		}
		kb.AddRow(row)
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrToToday), tg.NewCmd(CmdShowHome, nil)))

	err := b.showHTML(i18n.Tr("Repeat the task"), kb)
	if err != nil {
		return fmt.Errorf("showRecuringKeyboard : %w", err)
	}

	return nil
}

// addToFile adds content at the top of the file.
// Creates a file if not exists.
func (b *Bot) addToFile(dir, filename, content string) error {
	existingContent, err := b.fs.Read(dir, filename)
	// Ignore if file is missing, it would be created.
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("add to file: can't read existing file: %w", err)
	}

	header := fmt.Sprintf("#### %d %s %d, %s", now().Day(), now().Format("January"), now().Year(), now().Weekday())
	newContent := txt.AddHeaderAndText(existingContent, header, content)

	err = b.fs.Write(dir, filename, newContent)
	if err != nil {
		return fmt.Errorf("add to file: can't save file: %w", err)
	}

	return nil
}

func (b *Bot) openInApp(_ []string) error {
	token := sync.GenOneTimeToken(b.userID)
	onetimeURL := fmt.Sprintf("%s?token=%s", config.ServerCfg.AppURL, token)
	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr("Open in app"), tg.NewURLCmd(onetimeURL))})

	return b.showHTML(i18n.Tr("🔗 Here's your <b>one-time</b> link! <b>Desktop-only</b> for now."), kb)
}

func (b *Bot) showHelp(_ []string) error {
	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))})

	return b.showHTML(i18n.Tr("Refer to files.md for help!"), kb)
}

func (b *Bot) download(_ []string) error {
	kb := tg.NewKeyboard([]tg.Row{tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil))})

	return b.showHTML(i18n.Tr("Not yet implemented 🏗!"), kb)
}

func (b *Bot) setTasksOnlyMode(_ []string) error {
	err := b.cfg.SetMode(userconfig.ModeTasks)
	if err != nil {
		return fmt.Errorf("tasks only mode: can't set notes only mode %w", err)
	}

	cmds, err := b.cfg.MoveToCmds()
	if err != nil {
		return fmt.Errorf("tasks only mode: can't get quick commands %w", err)
	}

	for _, cmd := range cmds {
		err = b.cfg.DelMoveToCmd(cmd)
		if err != nil {
			return fmt.Errorf("tasks only mode: can't delete quick command %w", err)
		}
	}

	moveToCmds := []string{
		CmdScheduleForTmrw,
		CmdMoveToLater,
		CmdShowScheduleForDay,
	}
	for _, cmd := range moveToCmds {
		err = b.cfg.AddMoveToCmd(cmd)
		if err != nil {
			return fmt.Errorf("full mode: can't add quick command %w", err)
		}
	}

	return b.ShowHome(nil)
}

func (b *Bot) setNotesOnlyMode(_ []string) error {
	err := b.cfg.SetMode(userconfig.ModeNotes)
	if err != nil {
		return fmt.Errorf("notes only mode: can't set notes only mode %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) setJournalOnlyMode(_ []string) error {
	err := b.cfg.SetMode(userconfig.ModeJournal)
	if err != nil {
		return fmt.Errorf("journal only mode: can't set notes only mode %w", err)
	}

	return b.showHTML(i18n.Tr("What's on your mind?"), nil)
}

func (b *Bot) setFullMode(_ []string) error {
	err := b.cfg.SetMode(userconfig.ModeFull)
	if err != nil {
		return fmt.Errorf("full mode: can't set notes only mode %w", err)
	}

	moveToCmds := []string{
		CmdShowMoveToDirOrFile,
		CmdMoveToRead,
		CmdMoveToShop,
		CmdMoveToWatch,
		CmdMoveToJournal,
		CmdMoveToDraft,
		CmdMoveToFinalize,
		CmdMoveToDiscussion,
	}
	for _, cmd := range moveToCmds {
		err = b.cfg.AddMoveToCmd(cmd)
		if err != nil {
			return fmt.Errorf("full mode: can't add quick command %w", err)
		}
	}

	err = b.fs.CreateSystemDirs()
	if err != nil {
		return fmt.Errorf("full mode: can't create dirs: %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) setChatOnlyMode(_ []string) error {
	err := b.cfg.SetMode(userconfig.ModeChat)
	if err != nil {
		return fmt.Errorf("chat only mode: can't set chat only mode %w", err)
	}

	return b.showHTML(i18n.Tr("What's on your mind?"), nil)
}

func (b *Bot) completeHabit(params []string) error {
	habit := params[0]
	userHabits, err := habits.Habits(b.fs, time.Now().Year())
	if err != nil {
		return fmt.Errorf("complete habit: can't get habits: %w", err)
	}

	userHabits[habit][time.Now().YearDay()] = 1

	err = habits.Write(b.fs, time.Now().Year(), userHabits)
	if err != nil {
		return fmt.Errorf("complete habit: can't write habits: %w", err)
	}

	emoji := habits.Emoji(b.fs, habit)

	userConf := userconfig.NewConfig(b.fs, b.userID, config.ServerCfg.ConfigFilename)
	err = journal.AddEmoji(b.fs, emoji, userConf.Timezone())
	if err != nil {
		return fmt.Errorf("complete habit: can't write emoji to journal: %w", err)
	}

	record := fmt.Sprintf("%s %s", emoji, habit)
	err = journal.AddRecord(b.fs, record, userConf.Timezone(), userConf.JournalTimestampsEnabled())
	if err != nil {
		return fmt.Errorf("complete habit: can't write record to journal: %w", err)
	}

	return b.ShowHome(nil)
}

func (b *Bot) shareNote(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.Unhash(fs.DirUserRoot, dirHash)
	if err != nil {
		return fmt.Errorf("share note: can't find dir: %w", err)
	}

	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("share note: can't find file: %w", err)
	}

	content, err := b.fs.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("share note: %w", err)
	}

	for _, channel := range b.cfg.Channels() {
		probablyInvalidMD := fmt.Sprintf("**%s/%s**\n\n%s", fs.DisplayName(dir), fs.DisplayName(filename), content)
		probablyInvalidMD, images, _ := txt.ExtractTextImgsLinks(probablyInvalidMD)
		// Sending a gallery of images if there are any
		if len(images) > 0 {
			// We tolerate errors with the image gallery for now, text is more important
			mids, imgErr := b.tg.SendImages(channel, images)
			if imgErr == nil {
				for _, imgMid := range mids {
					b.db.AddImgMsgID(imgMid)
				}
			} else {
				slog.Error("Can't send images", "error", imgErr)
			}
		}

		// If our msg is too long, we send maxMsgsToSendAtOnce first messages.
		// Keyboard is attached to the last one
		textChunks := txt.SplitTextIntoChunks(probablyInvalidMD, maxMsgLength)
		textChunks = textChunks[0:min(maxMsgsToSendAtOnce, len(textChunks))]
		lastChunk := textChunks[len(textChunks)-1]
		textChunks = textChunks[0 : len(textChunks)-1]
		for _, textChunk := range textChunks {
			_, _ = b.tg.Send(b.userID, txt.MarkdownToHTML(textChunk), nil, tg.MarkupHTML)
		}

		_, err := b.tg.Send(channel, txt.MarkdownToHTML(lastChunk), nil, tg.MarkupHTML)
		if err != nil {
			return fmt.Errorf("share: %w", err)
		}
	}

	return nil
}

func extractMarkdown(u Update) string {
	content := txt.TelegramEntitiesToMarkdown(u.MsgText(), u.MsgEntities())
	content = strings.TrimSpace(txt.NormNewLines(content))

	return txt.Ucfirst(content)
}

func checklistTitle(checklist string) string {
	checklist = strings.TrimSuffix(checklist, filepath.Ext(checklist))
	stripChecklistChars := regexp.MustCompile(`^_.*?_(.+)`)
	title := stripChecklistChars.ReplaceAllString(checklist, "$1")
	title = strings.TrimPrefix(strings.TrimSuffix(title, "_"), "_")

	return fs.DisplayName(title)
}

func taskBtnLabel(task string) string {
	if task == fs.PomodoroTask {
		return i18n.AddEmoji(i18n.Tr(task))
	}
	return i18n.AddEmoji(task)
}

func completedMsg() string {
	msgs := []string{
		i18n.Tr("Completed! 🚀"),
		i18n.Tr("Done! 🎉"),
		i18n.Tr("Awesome! 💪"),
		i18n.Tr("Great job! 🌟"),
		i18n.Tr("Good work! 🎈"),
		i18n.Tr("Nice! 🎊"),
		i18n.Tr("Fantastic! 🎇"),
		i18n.Tr("Excellent! 🎯"),
		i18n.Tr("Perfect! 🏆"),
		i18n.Tr("Bravo! 👏"),
		i18n.Tr("Superb! 🌠"),
		i18n.Tr("You did it! ✅"),
		i18n.Tr("Nicely done! 🎖"),
		i18n.Tr("Nailed it! 🎯"),
	}

	return msgs[rand.Intn(len(msgs))]
}

func splitUserRelativePath(value string) (dir, filename string) {
	value = strings.TrimPrefix(filepath.ToSlash(value), "/")
	if !strings.Contains(value, "/") {
		return fs.DirUserRoot, value
	}
	i := strings.LastIndex(value, "/")
	return value[:i], value[i+1:]
}
