package ui

import (
	"fyne.io/fyne/v2/widget"
)

func NewLabeledEntry(label string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(label)
	return entry
}
