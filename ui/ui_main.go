package ui

import (
	"fmt"
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
)

func Init(a fyne.App, w fyne.Window, c backive.Configuration, d backive.Database) {
	app = a
	a.SetIcon(theme.FyneLogo())
	makeTray(app)
	config = c
	db = d
}
func NotificationRun() {
	if doNotShowUntil == time.Unix(0, 0) || time.Now().After(doNotShowUntil) {
		ShowNotification()
		if doNotShowUntil != time.Unix(0, 0) {
			doNotShowUntil = time.Unix(0, 0)
		}
	}
	h, _ := time.ParseDuration("15m")
	time.Sleep(h)
}

func ShowNotification() {
	displayStr, err := MakeNotificationString()
	if err == nil {
		app.SendNotification(
			fyne.NewNotification(
				"Backups are overdue...",
				fmt.Sprintf("Name\t(device)\t[overdue]\n%s", displayStr),
			),
		)
	}
}

func MakeNotificationString() (string, error) {
	db.Load()
	var displayStr string = ""
	var runs backive.Runs
	runs.Load(db)
	fmt.Printf("Notification run\n")
	for _, v := range config.Backups {
		fmt.Printf("Notification run %s\n", v.Name)
		if v.ShouldRun() && v.Frequency > 0 {
			fmt.Printf("Notification for %s\n", v.Name)
			lastBackup, err := runs.LastRun(v)
			if err != nil {
				return "", err
			}
			freq, _ := time.ParseDuration(fmt.Sprintf("%dd", v.Frequency))
			days := time.Now().Sub(lastBackup.Add(freq))
			displayStr += fmt.Sprintf("%s\t(%s)\t[%f days]\n", v.Name, v.TargetDevice, days.Hours()/24)
		}
	}
	return displayStr, nil
}

func makeTray(app fyne.App) {
	if desk, ok := app.(desktop.App); ok {
		menu := fyne.NewMenu(
			"backive",
			fyne.NewMenuItem("Show notifications again", func() {
				ShowNotification()
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
