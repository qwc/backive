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
	app := app.New()
	w := app.NewWindow("Hello World!")
	backiveui.Init(app, w, config, database)

	//w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}
