package server

import (
	"fmt"

	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/pkg/tg"
)

const (
	addBtn = "➕"
	delBtn = "➖"
)

var AvailableMoveToBtns = []tg.Btn{
	tg.NewBtn(i18n.Tr(i18n.StrToTomorrow), tg.NewCmd(CmdScheduleForTmrw, nil)),
	tg.NewBtn(i18n.Tr(i18n.StrToLater), tg.NewCmd(CmdMoveToLater, nil)),
	tg.NewBtn(i18n.Tr(i18n.StrToDraft), tg.NewCmd(CmdMoveToDraft, nil)),
	tg.NewBtn(i18n.Tr(i18n.StrToFinalize), tg.NewCmd(CmdMoveToFinalize, nil)),
	tg.NewBtn(i18n.Tr(i18n.StrToDiscussion), tg.NewCmd(CmdMoveToDiscussion, nil)),
}

var AvailableQuickBtns = []tg.Btn{}

func (b *Bot) showSettings(_ []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) showTimezone(_ []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) setTimezone(params []string) error {
	timezone := params[0]
	err := b.cfg.SetTimezone(timezone)
	if err != nil {
		return fmt.Errorf("setTimezone : %w", err)
	}
	return b.ShowHome(nil)
}

func (b *Bot) showQuickBtnsSettings(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) addToQuickBtns(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) delFromQuickBtns(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) quickBtns() []tg.Btn {
	return nil
}

func (b *Bot) showMoveToBtnsSettings(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) addToMoveToBtns(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) delFromMoveToBtns(params []string) error {
	return b.ShowHome(nil)
}

func (b *Bot) moveToBtns(msgHash string) []tg.Btn {
	moveToBtns := tg.NewRow()
	cmds, err := b.cfg.MoveToCmds()
	if err != nil {
		return nil
	}
	for _, cmd := range cmds {
		for _, btn := range AvailableMoveToBtns {
			if btn.Cmd.Name == cmd {
				btn.Cmd.Params = []string{msgHash}
				moveToBtns = append(moveToBtns, btn)
				break
			}
		}
	}
	return moveToBtns
}
