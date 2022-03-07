package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/qwc/backive"
)

var (
	app    fyne.App
	window fyne.Window
	config backive.Configuration
	db     backive.Database

	accord *widget.Accordion
	center *fyne.Container
)

func Init(a fyne.App, w fyne.Window, c backive.Configuration, d backive.Database) {
	app = a
	window = w
	config = c
	db = d
	SetupLayout()
}

func SetupLayout() {
	accord = widget.NewAccordion()
	center := container.NewMax()
	content := container.New(layout.NewBorderLayout(nil, nil, accord, nil), accord, center)
	window.SetContent(content)
}
