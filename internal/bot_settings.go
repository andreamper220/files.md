package internal

import (
	"fmt"

	"zakirullin/stuffbot/config"
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/consts"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/tg"
	"zakirullin/stuffbot/pkg/txt"
)

const (
	addBtn = "➕"
	delBtn = "➖"
)

func (b *Bot) showSettings(params []string) error {
	var kb tg.Keyboard
	kb.AddRow(tg.NewBtn(i18n.StrQuickBtns, tg.NewCmd(consts.CmdShowQuickBtnsSettings, nil)))
	kb.AddRow(tg.NewBtn(i18n.StrMoveToBtns, tg.NewCmd(consts.CmdShowMoveToBtnsSettings, nil)))
	kb.AddRow(tg.NewBtn(i18n.StrToday, tg.NewCmd(consts.CmdShowToday, nil)))

	err := b.show("Settings:", &kb, tg.MarkupHTML)
	if err != nil {
		return fmt.Errorf("showSettings : %w", err)
	}

	return nil
}

func (b *Bot) showQuickBtnsSettings(params []string) error {
	var kb tg.Keyboard

	// Step 1. Append all buttons that are already chosen by user
	var usedCmds []string

	// We iterate through hardcoded panel to preserve order of buttons in UI
	cmds, err := b.conf.QuickCmds()
	if err != nil {
		return fmt.Errorf("can't get quick cmds: %w", err)
	}

	for _, cmd := range cmds {
		for _, btn := range userconfig.AvailableQuickBtns {
			if btn.Cmd.Name != cmd {
				continue
			}

			name := fmt.Sprintf("%s %s %s", i18n.Emoji(btn.Name), btn.Name, delBtn)
			enabledCmd := tg.NewCmd(consts.CmdDelFromQuickBtns, []string{btn.Cmd.Name})
			kb.AddRow(tg.NewBtn(name, enabledCmd))
			usedCmds = append(usedCmds, cmd)
			break
		}
	}

	kb.AddRow(tg.NewBtn("-", tg.NewCmd(consts.CmdDoNothing, nil)))

	// Step 2. now, let's fill buttons that are not disabled...
	for _, btn := range userconfig.AvailableQuickBtns {
		// Check if command is enabled
		cmdUsed := false
		for _, usedCmd := range usedCmds {
			if btn.Cmd.Name == usedCmd {
				cmdUsed = true
			}
		}
		if cmdUsed {
			continue
		}
		// Command is not enabled, so add it to disabled list
		name := fmt.Sprintf("%s %s %s", i18n.Emoji(btn.Name), btn.Name, addBtn)
		disabledCmd := tg.NewCmd(consts.CmdAddToQuickBtns, []string{btn.Cmd.Name})
		kb.AddRow(tg.NewBtn(name, disabledCmd))
	}

	kb.AddRow(tg.NewBtn(i18n.StrBack, tg.NewCmd(consts.CmdShowSettings, nil)))

	text := fmt.Sprintf("Configure quick buttons (%s = add to quick buttons, %s = to remove from quick buttons):", addBtn, delBtn)
	err = b.show(text, &kb, tg.MarkupHTML)
	if err != nil {
		return fmt.Errorf("configureQuickPanel : %w", err)
	}

	return nil
}

func (b *Bot) addToQuickBtns(params []string) error {
	cmd := params[0]

	// Search whether a command is valid
	found := false
	for _, btn := range userconfig.AvailableQuickBtns {
		if btn.Cmd.Name == cmd {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unknown command: %s", cmd)
	}

	err := b.conf.AddQuickCmd(cmd)
	if err != nil {
		return fmt.Errorf("can't add to quick buttons: %w", err)
	}

	return b.showQuickBtnsSettings([]string{})
}

func (b *Bot) delFromQuickBtns(params []string) error {
	cmd := params[0]

	_ = b.conf.DelQuickCmd(cmd)

	return b.showQuickBtnsSettings([]string{})
}

func (b *Bot) quickBtns() []tg.Btn {
	quickBtnsRow := tg.NewRow()
	// We can tolerate missing quick btns
	cmds, _ := b.conf.QuickCmds()
	for _, cmd := range cmds {
		for _, btn := range userconfig.AvailableQuickBtns {
			if btn.Cmd.Name == cmd {
				if btn.Cmd.Name == consts.CmdWebAppHabits {
					habitsUrl := fmt.Sprintf("%s/habits_v2/%d", config.Config.Host, b.userID)
					btn.Cmd.Params = []string{habitsUrl}
				}
				btn.Name = i18n.Emoji(btn.Name)

				quickBtnsRow = append(quickBtnsRow, btn)
				break
			}
		}
	}

	return quickBtnsRow
}

// A little copy-paste from showQuickBtnsSettings
func (b *Bot) showMoveToBtnsSettings(params []string) error {
	var kb tg.Keyboard

	// Step 1. Append all buttons that are already chosen by user
	var usedCmds []string

	// We iterate through hardcoded panel to preserve order of buttons in UI
	for _, cmd := range b.conf.MoveToCmds() {
		for _, btn := range userconfig.AvailableMoveToBtns {
			if btn.Cmd.Name != cmd {
				continue
			}

			name := txt.Emoji(delBtn, btn.Name)
			enabledCmd := tg.NewCmd(consts.CmdDelFromMoveToBtns, []string{btn.Cmd.Name})
			kb.AddRow(tg.NewBtn(name, enabledCmd))
			usedCmds = append(usedCmds, cmd)
			break
		}
	}

	kb.AddRow(tg.NewBtn("-", tg.NewCmd(consts.CmdDoNothing, nil)))

	// Step 2. now, let's fill buttons that are not disabled...
	for _, btn := range userconfig.AvailableMoveToBtns {
		// Check if command is enabled
		cmdUsed := false
		for _, usedCmd := range usedCmds {
			if btn.Cmd.Name == usedCmd {
				cmdUsed = true
			}
		}
		if cmdUsed {
			continue
		}
		// Command is not enabled, so add it to disabled list
		name := txt.Emoji(addBtn, btn.Name)
		disabledCmd := tg.NewCmd(consts.CmdAddToMoveToBtns, []string{btn.Cmd.Name})
		kb.AddRow(tg.NewBtn(name, disabledCmd))
	}

	kb.AddRow(tg.NewBtn(i18n.StrBack, tg.NewCmd(consts.CmdShowSettings, nil)))

	text := fmt.Sprintf("Configure quick panel (%s = add to panel, %s = to remove):", addBtn, delBtn)
	err := b.show(text, &kb, tg.MarkupHTML)
	if err != nil {
		return fmt.Errorf("configureQuickPanel : %w", err)
	}

	return nil
}

func (b *Bot) addToMoveToBtns(params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("no params suplied to addToPanel")
	}

	// Search whether a command is valid
	found := false
	for _, btn := range userconfig.AvailableMoveToBtns {
		if btn.Cmd.Name == params[0] {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unknown command: %s", params[0])
	}

	if !b.conf.AddMoveToCmd(params[0]) {
		return fmt.Errorf("button already exists in user's prefs: %s", params[0])
	}

	return b.showMoveToBtnsSettings([]string{})
}

func (b *Bot) delFromMoveToBtns(params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("no params suplied to delFromPanel")
	}
	if !b.conf.DelMoveToCmd(params[0]) {
		return fmt.Errorf("button doesn't exist in user's prefs: %s", params[0])
	}

	return b.showMoveToBtnsSettings([]string{})
}

func (b *Bot) moveToBtns(filenameHash string) []tg.Btn {
	moveToBtns := tg.NewRow()

	// We iterate through hardcoded panel to preserve order of buttons in UI
	for _, cmd := range b.conf.MoveToCmds() {
		for _, btn := range userconfig.AvailableMoveToBtns {
			if btn.Cmd.Name == cmd {
				btn.Cmd.Params = []string{filenameHash}
				moveToBtns = append(moveToBtns, btn)
				break
			}
		}
	}

	return moveToBtns
}
