package userconfig

import (
	"fmt"

	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/consts"
	"zakirullin/stuffbot/pkg/tg"
)

var AvailableMoveToBtns = []tg.Btn{
	tg.NewBtn(i18n.StrToTomorrow, tg.NewCmd(consts.CmdScheduleForTmrw, nil)),
	tg.NewBtn(i18n.StrMoveToLater, tg.NewCmd(consts.CmdMoveToLater, nil)),
	tg.NewBtn(i18n.StrToADay, tg.NewCmd(consts.CmdShowScheduleForDay, nil)),
	tg.NewBtn(i18n.StrToFile, tg.NewCmd(consts.CmdShowMoveToFile, nil)),
	tg.NewBtn(i18n.StrToJournal, tg.NewCmd(consts.CmdMoveToJournal, nil)),
	tg.NewBtn(i18n.StrToRead, tg.NewCmd(consts.CmdMoveToRead, nil)),
	tg.NewBtn(i18n.StrToWatch, tg.NewCmd(consts.CmdMoveToWatch, nil)),
	tg.NewBtn(i18n.StrToShop, tg.NewCmd(consts.CmdMoveToShop, nil)),
	tg.NewBtn(i18n.StrToChecklist, tg.NewCmd(consts.CmdShowMoveToChecklist, nil)),
}

func (c *Config) AddMoveToCmd(cmd string) error {
	lock := c.userLock()
	lock.Lock()
	defer lock.Unlock()

	conf, err := c.read(c.path)
	if err != nil {
		return fmt.Errorf("can't add move to cmd: can't read config: %w", err)
	}
	conf.MoveToCmds = append(conf.MoveToCmds, cmd)
	err = c.write(c.path, conf)
	if err != nil {
		return fmt.Errorf("can't add move to cmd: can't write config: %w", err)
	}

	return nil
}

func (c *Config) MoveToCmds() ([]string, error) {
	conf, err := c.read(c.path)
	if err != nil {
		return nil, fmt.Errorf("can't get move to cmds: can't read config: %w", err)
	}

	return conf.MoveToCmds, nil
}

func (c *Config) DelMoveToCmd(cmd string) error {
	lock := c.userLock()
	lock.Lock()
	defer lock.Unlock()

	conf, err := c.read(c.path)
	if err != nil {
		return fmt.Errorf("can't del move to cmd: can't read config: %w", err)
	}

	var newCmds []string
	for _, curMoveToCmd := range conf.MoveToCmds {
		if curMoveToCmd != cmd {
			newCmds = append(newCmds, curMoveToCmd)
		}
	}
	conf.MoveToCmds = newCmds

	err = c.write(c.path, conf)
	if err != nil {
		return fmt.Errorf("can't del move to cmd: can't write config: %w", err)
	}

	return nil
}
