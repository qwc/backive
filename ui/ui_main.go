package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/qwc/backive"
)

var mockOsWriteFile = os.WriteFile
var mockOsReadFile = os.ReadFile

type UISettings struct {
	hideUntil   time.Time
	globalLevel int
}

var (
	app            fyne.App
	window         fyne.Window
	config         backive.Configuration
	doNotShowUntil time.Time = time.Unix(0, 0)
	c              net.Conn
	uisettings     UISettings
	messageLevel   int
	apphomedir     string
)

func Init(a fyne.App, w fyne.Window, conf backive.Configuration) {
	app = a
	a.SetIcon(theme.FyneLogo())
	makeTray(app)
	config = conf
	apphomedir, _ := os.UserHomeDir()
	apphomedir += string(os.PathSeparator) + ".config" + string(os.PathSeparator) + "backive" + string(os.PathSeparator) + "ui.json"
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
	level, err := strconv.ParseUint(data["level"], 10, 64)
	if err != nil {
		return false
	}
	if level <= 10 {
		return true
	}
	if int(level) <= uisettings.globalLevel && messageLevel > 0 {
		return false
	}
	if int(level) <= uisettings.globalLevel {
		return true
	}
	return false
}

func SetHideUntil(until time.Time) {
	uisettings.hideUntil = until
}

func SetMessageLevel(level int) {
	if level <= 10 {
		messageLevel = level
	} else {
		uisettings.globalLevel = level
		messageLevel = 0
		SaveSettings()
	}
}

func SaveSettings() {
	// save internal settings to file in homedir
	jsonstr, merr := json.Marshal(uisettings)
	if merr != nil {
		panic(merr)
	}
	log.Printf("Writing database output to file: %s", jsonstr)
	saveDir, _ := path.Split(apphomedir)
	backive.CreateDirectoryIfNotExists(saveDir)
	err := mockOsWriteFile(apphomedir, []byte(jsonstr), 0644)
	if err != nil {
		panic(err)
	}
}

func LoadSettings() {
	// load settings
	if _, err := os.Stat(apphomedir); err == nil {
		data, rferr := mockOsReadFile(apphomedir)
		if rferr != nil {
			panic(rferr)
		}
		json.Unmarshal(data, &uisettings)
	} /*else if os.IsNotExist(err) {
		// no data
	}*/
}

func makeTray(app fyne.App) {
	if desk, ok := app.(desktop.App); ok {
		hideReminders := fyne.NewMenuItem(
			"Hide Reminders for",
			nil,
		)
		hideReminders.ChildMenu = fyne.NewMenu(
			"",
			fyne.NewMenuItem("Hide reminder notifications for today", func() {
				doNotShowUntil = time.Now().AddDate(0, 0, 1)
				SetHideUntil(doNotShowUntil)
			}),
			fyne.NewMenuItem("Hide reminder notifications for a hour", func() {
				doNotShowUntil = time.Now().Add(time.Hour)
				SetHideUntil(doNotShowUntil)
			}),
		)
		levelMenu := fyne.NewMenuItem(
			"Set notification level", nil,
		)
		levelMenu.ChildMenu = fyne.NewMenu(
			"",
			fyne.NewMenuItem(
				"Only problems and tasks finished (resets to previous with restart)",
				func() { SetMessageLevel(10) },
			),
			fyne.NewMenuItem(
				"+ Reminders (default)",
				func() { SetMessageLevel(20) },
			),
			fyne.NewMenuItem(
				"+ Informational messages",
				func() { SetMessageLevel(30) },
			),
			fyne.NewMenuItem(
				"+ Verbose/Debug messages",
				func() { SetMessageLevel(40) },
			),
		)
		menu := fyne.NewMenu(
			"backive",
			hideReminders,
			levelMenu,
		)
		desk.SetSystemTrayMenu(menu)
	}
}
