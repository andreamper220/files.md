package userconfig

import (
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/constants"
	"zakirullin/stuffbot/pkg/tg"
)

type QuickPanelBtn struct {
	Cmd         string
	CmdType     string
	Emoji       string
	Description string
}

var QuickPanelAvailableBtns = []QuickPanelBtn{
	NewQuickPanelBtn(constants.CmdShowDoc, tg.CallbackCmd, i18n.EmDocs, "Documents"),
	NewQuickPanelBtn(constants.CmdShowChecklists, tg.CallbackCmd, i18n.EmCheckList, "Checklists"),
	NewQuickPanelBtn(constants.CmdShowPostpone, tg.CallbackCmd, i18n.EmPostpone, "Postpone"),
	NewQuickPanelBtn(constants.CmdShowReadChecklist, tg.CallbackCmd, i18n.Emoji("Read"), "Read"),
	NewQuickPanelBtn(constants.CmdShowWatchChecklist, tg.CallbackCmd, i18n.Emoji("Watch"), "Watch"),
	NewQuickPanelBtn(constants.CmdShowShopChecklist, tg.CallbackCmd, i18n.Emoji("Shop"), "Shop"),
}

func NewQuickPanelBtn(cmd, cmdType, emoji, description string) QuickPanelBtn {
	return QuickPanelBtn{cmd, cmdType, emoji, description}
}