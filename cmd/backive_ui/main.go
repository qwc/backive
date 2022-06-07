package main

import (
	"time"

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
	backive.Init(config, database)
	app := app.NewWithID("Backive UI")
	backiveui.Init(app, nil, config, database)
	go func() {
		for {
			backiveui.NotificationRun()
			time.Sleep(time.Second)
		}
	}()

	app.Run()
}
