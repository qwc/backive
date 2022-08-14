package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/qwc/backive"
	backiveui "github.com/qwc/backive/ui"
)

var (
	config backive.Configuration
)

func main() {

	config.Load()
	app := app.NewWithID("Backive UI")
	backiveui.Init(app, nil, config)
	app.Run()
}
