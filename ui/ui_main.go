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

type uiSettings struct {
	hideUntil   time.Time
	globalLevel int
}

var (
	app            fyne.App
	window         fyne.Window
	config         backive.Configuration
	doNotShowUntil time.Time = time.Unix(0, 0)
	c              net.Conn
	uisettings     uiSettings
	messageLevel   int
	apphomedir     string
)

// Init the fyne application
func Init(a fyne.App, w fyne.Window, conf backive.Configuration) {
	app = a
	a.SetIcon(theme.FyneLogo())
	makeTray(app)
	config = conf
	apphomedir, _ := os.UserHomeDir()
	apphomedir += string(os.PathSeparator) + ".config" + string(os.PathSeparator) + "backive" + string(os.PathSeparator) + "ui.json"
	LoadSettings()
	go PollConnection()
	fmt.Println("UI started")
}

// PollConnection polls in an endless loop the connection
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

// ShowNotification shows a single notification
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

// ShallShow checks if a message should be shown by level
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

// SetHideUntil sets the time until messages should be hidden
func SetHideUntil(until time.Time) {
	uisettings.hideUntil = until
	SaveSettings()
}

// SetMessageLevel does exactly that.
func SetMessageLevel(level int) {
	if level <= 10 {
		messageLevel = level
	} else {
		uisettings.globalLevel = level
		messageLevel = 0
		SaveSettings()
	}
}

// SaveSettings stores the settings in $HOME/.config/backive/ui.json
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
	fmt.Println("Saved Settings.")
}

// LoadSettings loads the settings from the place where SaveSettings stored them.
func LoadSettings() {
	// load settings
	if _, err := os.Stat(apphomedir); err == nil {
		data, rferr := mockOsReadFile(apphomedir)
		if rferr != nil {
			panic(rferr)
		}
		json.Unmarshal(data, &uisettings)
		fmt.Println("Loaded Settings.")
	} /*else if os.IsNotExist(err) {
		// no data
	}*/
}

// makeTray creates the tray menus needed.
func makeTray(app fyne.App) {
	if desk, ok := app.(desktop.App); ok {
		hideReminders := fyne.NewMenuItem(
			"Hide reminders for",
			nil,
		)
		hideReminders.ChildMenu = fyne.NewMenu(
			"",
			fyne.NewMenuItem("today", func() {
				doNotShowUntil = time.Now().AddDate(0, 0, 1)
				SetHideUntil(doNotShowUntil)
			}),
			fyne.NewMenuItem("a hour", func() {
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
