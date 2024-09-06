package gui

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"zakirullin/stuffbot/internal"
	"zakirullin/stuffbot/pkg/tg"
)

type Chat struct {
	userID   int64
	input    *widget.Entry
	messages *fyne.Container
	updater  func(updInterface internal.UpdInterface) error
}

func NewChat(userID int64, updater func(u internal.UpdInterface) error) *Chat {
	return &Chat{userID: userID, updater: updater, input: widget.NewEntry(), messages: container.NewVBox()}
}

func (c *Chat) Run(startupCMD tg.Cmd) {
	a := app.New()
	w := a.NewWindow("Files.md")

	c.input.OnSubmitted = func(msg string) {
		//c.input.MultiLine = true
		c.Send(c.userID, msg, nil, "")
		c.updater(tg.NewFakeUpd(1, msg))
	}

	sendBtn := widget.NewButton("✉️", func() {
		//c.input.MultiLine = false
		msg := c.input.Text
		c.Send(c.userID, msg, nil, "")
		c.updater(tg.NewFakeUpd(1, msg))
	})

	// Make sure the input field takes all available width
	inputLine := container.New(layout.NewBorderLayout(nil, nil, nil, sendBtn), c.input, sendBtn)

	// Container with message and input line
	box := container.NewVBox(c.messages, inputLine)
	cont := container.New(layout.NewBorderLayout(nil, box, nil, nil), box)
	w.SetContent(cont)

	w.Resize(fyne.NewSize(400, 300)) // Set initial window size
	w.Canvas().Focus(c.input)
	w.Show()
	c.updater(tg.NewFakeUpdCmd(1, startupCMD))
	a.Run()
}

func (c *Chat) Send(userID int64, text string, kb *tg.Keyboard, markup string) (int, error) {
	if len(text) == 0 {
		return 0, nil
	}

	msgContainer := container.NewVBox(
		widget.NewLabel(text),
	)
	c.attachKeyboard(kb, msgContainer)

	c.input.SetText("")
	c.messages.Add(msgContainer)

	return 0, nil
}

func (c *Chat) Edit(userID int64, msgID int, text string, kb *tg.Keyboard, markup string) error {
	if len(text) == 0 {
		return nil
	}

	c.messages.RemoveAll()
	c.Send(userID, text, kb, markup)

	return nil
}

func (c *Chat) Del(userID int64, msgID int) error {
	return nil
}

func (c *Chat) AnswerCallbackQuery(queryID string, text string) error {
	return nil
}

func (c *Chat) AnswerInlineQuery(queryID string, results []interface{}, cacheTime int, offset string) error {
	return nil
}

func (c *Chat) DownloadFile(fileID string, outFile io.Writer) (string, error) {
	return "", nil
}

func (c *Chat) attachKeyboard(kb *tg.Keyboard, msgContainer *fyne.Container) {
	btnCallback := func(cmd tg.Cmd) func() {
		return func() {
			c.updater(tg.NewFakeUpdCmd(1, cmd))
		}
	}

	var kbBtns []tg.Row
	if kb != nil {
		kbBtns = kb.Btns
	}
	for _, row := range kbBtns {
		switch row.(type) {
		case tg.Btn:
			b := row.(tg.Btn)
			msgContainer.Add(widget.NewButton(b.Name, btnCallback(b.Cmd)))
		case []tg.Btn:
			btns := row.([]tg.Btn)
			rowContainer := container.New(layout.NewGridLayoutWithColumns(len(btns)))
			for _, b := range btns {
				rowContainer.Add(widget.NewButton(b.Name, btnCallback(b.Cmd)))
			}
			msgContainer.Add(rowContainer)
		}
	}
}
