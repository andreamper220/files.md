package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/morningsummary"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/priority"
	"github.com/zakirullin/files.md/server/userconfig"
)

func emptyHomeText() string {
	return "🌴 " + i18n.Tr("Nothing here yet - send me something!")
}

func homeMessageText(userFS *fs.FS, cfg *userconfig.Config, shownCount int) string {
	label := emptyHomeText()
	if shownCount > 0 {
		label = homeCountText(shownCount)
	}

	report, err := morningsummary.Build(userFS, cfg)
	if err != nil {
		return label
	}
	report = strings.TrimSpace(report)
	if report == "" {
		return label
	}
	return report + "\n\n" + label
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
