package ui

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/qwc/backive"
)

var (
	app            fyne.App
	window         fyne.Window
	config         backive.Configuration
	db             backive.Database
	doNotShowUntil time.Time = time.Unix(0, 0)
	c net.Conn
)

func Init(a fyne.App, w fyne.Window, conf backive.Configuration, d backive.Database) {
	app = a
	a.SetIcon(theme.FyneLogo())
	makeTray(app)
	config = conf
	db = d
	go PollConnection()
}

func PollConnection() {
	var err error
	for {
		if c == nil {
			c, err = net.Dial("unix", config.Settings.UIUnixSocketLocation)
		} else {
			err = fmt.Errorf("Connection already established")
		}
		if err != nil {
			// ignore
			err = nil
			// sleep a while and then retry
			time.Sleep(10 * time.Second)
		}
	}
}

func NotificationRun() {
	if c != nil {
		b := make([]byte, 2048)
		i, err := c.Read(b)
		if err == nil && i > 0 {
			var data map[string]string
			err = json.Unmarshal(b, &data)
			if err == nil {
				ShowNotification(data)
			}
			// else ignore and try to read again
			err = nil
		}
		// we just try again and discard the error
		err = nil
	}
	/*
	if doNotShowUntil == time.Unix(0, 0) || time.Now().After(doNotShowUntil) {
		ShowNotification()
		if doNotShowUntil != time.Unix(0, 0) {
			doNotShowUntil = time.Unix(0, 0)
		}
	}
	h, _ := time.ParseDuration("15m")
	time.Sleep(h)
	//*/
}

func ShowNotification(data map[string]string) {
	if ShallShow(data) {
		app.SendNotification(
			fyne.NewNotification(
				data["header"],
				data["message"],
			),
		)
	}
}

func ShallShow(data map[string]string) bool {
	return true
}

func makeTray(app fyne.App) {
	if desk, ok := app.(desktop.App); ok {
		menu := fyne.NewMenu(
			"backive",
			fyne.NewMenuItem("Show notifications again", func() {
			}),
			fyne.NewMenuItem("Hide notifications for today", func() {
				doNotShowUntil = time.Now().AddDate(0, 0, 1)
			}),
			fyne.NewMenuItem("Hide notifications for a hour", func() {
				doNotShowUntil = time.Now().Add(time.Hour)
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}
