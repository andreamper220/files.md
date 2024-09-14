package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// entry with multiline support
type entry struct {
	widget.Entry
	shiftHeld   bool
	rowsVisible int
}

func newEntry() *entry {
	e := &entry{}
	e.rowsVisible = 3
	e.Scroll = container.ScrollVerticalOnly
	e.PlaceHolder = "Type here..."
	e.ExtendBaseWidget(e)

	return e
}

func (e *entry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn {
		if e.shiftHeld && !e.MultiLine {
			e.MultiLine = true
			e.SetMinRowsVisible(e.rowsVisible)
			e.Wrapping = fyne.TextWrapBreak
			e.rowsVisible++
		} else if e.shiftHeld {
			e.rowsVisible++
			e.SetMinRowsVisible(e.rowsVisible)
		} else if !e.shiftHeld && e.MultiLine {
			e.Resize(fyne.NewSize(e.Size().Width, e.Size().Height/2))
			e.MultiLine = false
			e.rowsVisible = 2
		}

		// User is submitting the entry
		if !e.shiftHeld {
			sendMsg()
		}
	}

	e.Entry.TypedKey(key)
}

func (e *entry) KeyDown(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		e.shiftHeld = true
	}

	e.Entry.KeyDown(key)
}

func (e *entry) KeyUp(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		e.shiftHeld = false
	}

	e.Entry.KeyUp(key)
}
