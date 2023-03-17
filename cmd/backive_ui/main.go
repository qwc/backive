package main

import (
	"log"
	"os"
	"path"

	"fyne.io/fyne/v2/app"

	"github.com/qwc/backive"
	backiveui "github.com/qwc/backive/ui"
)

var (
	config backive.Configuration
)

func setupLogging() {
	apphomedir, _ := os.UserHomeDir()
	apphomedir = path.Join(apphomedir, ".config", "backive")
	logname := path.Join(apphomedir, "backiveui.log")
	logdir, _ := path.Split(logname)
	backive.CreateDirectoryIfNotExists(logdir)
	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
		panic("no logfile no info")
	}
	log.SetOutput(logfile)
	log.Println("Logging initialized")
}

func main() {

	config.Load()
	setupLogging()
	app := app.NewWithID("Backive UI")
	backiveui.Init(app, nil, config)
	app.Run()
}
