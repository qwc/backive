package ui

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

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

	accord     *widget.Accordion
	center     *fyne.Container
	content    *fyne.Container
	devBtnList []*widget.Button
	bacBtnList []*widget.Button
)

func Init(a fyne.App, w fyne.Window, c backive.Configuration, d backive.Database) {
	app = a
	window = w
	config = c
	db = d
	SetupLayout()
}

func SetupLayout() {
	fmt.Print("Setting up layout\n")
	accord = widget.NewAccordion()
	devBtnList := []*widget.Button{}
	devLayout := container.NewVBox()
	for _, obj := range config.Devices {
		btn := widget.NewButton(obj.Name, nil)
		devBtnList = append(devBtnList, btn)
		devLayout.Add(btn)
	}
	for _, obj := range devBtnList {
		dev := obj.Text
		fmt.Printf("Setting OnTapped on %s\n", dev)
		obj.OnTapped = func() {
			fmt.Printf("Btn %s\n", dev)
			DisplayDevice(dev)
		}
	}
	devices := widget.NewAccordionItem("Devices", devLayout)
	accord.Append(devices)
	bacBtnList := []*widget.Button{}
	bacLayout := container.NewVBox()
	for _, obj := range config.Backups {
		bkp := obj.Name
		btn := widget.NewButton(
			bkp, func() {
				DisplayBackup(bkp)
			})
		bacBtnList = append(bacBtnList, btn)
		bacLayout.Add(btn)
	}
	backups := widget.NewAccordionItem("Backups", bacLayout)
	accord.Append(backups)
	center = container.NewMax()
	left := container.NewMax()
	left.Add(accord)
	window.Resize(fyne.NewSize(800, 600))
	content = container.NewBorder(nil, nil, left, nil, center)
	window.SetContent(content)
}

func ClearCenter() {
	fmt.Print("ClearCenter\n")
	if center != nil && center.Objects != nil && len(center.Objects) > 0 {
		center.Objects = nil
		center.Refresh()
	}
	content.Refresh()
}

func DisplayDevice(dev string) {
	ClearCenter()
	vbox := container.NewVBox()
	dataForm := container.New(layout.NewFormLayout())
	vbox.Add(dataForm)
	device := config.Devices[dev]
	dataForm.Add(widget.NewLabel("Name"))
	dataForm.Add(widget.NewLabel(device.Name))
	dataForm.Add(widget.NewLabel("UUID"))
	dataForm.Add(widget.NewLabel(device.UUID))
	dataForm.Add(widget.NewLabel("Owner"))
	dataForm.Add(widget.NewLabel(device.OwnerUser))
	fmt.Printf("Adding device %s\n", dev)
	dataForm.Add(widget.NewLabel("Assigned backup"))
	var backups []string
	for _, obj := range config.Backups {
		if dev == obj.TargetDevice {
			backups = append(backups, obj.Name)
		}
	}
	dataForm.Add(widget.NewLabel(strings.Join(backups, "\n")))
	center.Add(vbox)
}

func DisplayBackup(bac string) {
	ClearCenter()
	dataForm := container.New(layout.NewFormLayout())
	backup := config.Backups[bac]
	dataForm.Add(widget.NewLabel("Name"))
	dataForm.Add(widget.NewLabel(backup.Name))
	dataForm.Add(widget.NewLabel("Frequency (days)"))
	dataForm.Add(widget.NewLabel(fmt.Sprintf("%d", backup.Frequency)))
	dataForm.Add(widget.NewLabel("Target device"))
	dataForm.Add(widget.NewLabel(backup.TargetDevice))
	dataForm.Add(widget.NewLabel("Target directory"))
	dataForm.Add(widget.NewLabel(backup.TargetPath))
	dataForm.Add(widget.NewLabel("Source path"))
	dataForm.Add(widget.NewLabel(backup.SourcePath))
	dataForm.Add(widget.NewLabel("Script to execute"))
	var scriptWArgs []string
	switch slice := backup.ScriptPath.(type) {
	case []interface{}:
		for _, v := range slice {
			scriptWArgs = append(scriptWArgs, v.(string))
		}
	case []string:
		for _, v := range slice {
			scriptWArgs = append(scriptWArgs, v)
		}
	case string:
		scriptWArgs = append(scriptWArgs, slice)
	}
	dataForm.Add(widget.NewLabel(strings.Join(scriptWArgs, " ")))

	dataForm.Add(widget.NewLabel("Executing user"))
	dataForm.Add(widget.NewLabel(backup.ExeUser))
	dataForm.Add(widget.NewLabel("Label"))
	dataForm.Add(widget.NewLabel(backup.Label))
	logEntry := widget.NewMultiLineEntry()
	content, err := ioutil.ReadFile(path.Join(config.Settings.LogLocation, bac+".log"))
	if err != nil {
		logEntry.SetText("Reading file failed")
	}
	logEntry.Disable()
	logEntry.SetText(string(content))
	vbox := container.NewBorder(dataForm, nil, nil, nil, logEntry)
	fmt.Printf("Adding backu %s\n", bac)
	center.Add(container.NewMax(vbox))
}
