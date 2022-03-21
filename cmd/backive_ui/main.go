package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/qwc/backive"
	backiveui "github.com/qwc/backive/ui"
)

var (
	config   backive.Configuration
	database backive.Database
)

func main() {

	database.Load()
	config.Load()
	app := app.NewWithID("Backive UI")
	w := app.NewWindow("Backive UI")
	backiveui.Init(app, w, config, database)

	//w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}
