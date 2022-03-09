package ui

import (
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
	accord = widget.NewAccordion()
	devBtnList := []*widget.Button{}
	devLayout := container.NewVBox()
	for _, obj := range config.Devices {
		btn := widget.NewButton(obj.Name, nil)
		devBtnList = append(devBtnList, btn)
		devLayout.Add(btn)
	}
	devices := widget.NewAccordionItem("Devices", devLayout)
	accord.Append(devices)
	bacBtnList := []*widget.Button{}
	bacLayout := container.NewVBox()
	for _, obj := range config.Backups {
		btn := widget.NewButton(obj.Name, nil)
		bacBtnList = append(bacBtnList, btn)
		bacLayout.Add(btn)
	}
	backups := widget.NewAccordionItem("Backups", bacLayout)
	accord.Append(backups)
	center := container.NewMax()
	left := container.NewMax()
	left.Add(accord)
	window.Resize(fyne.NewSize(800, 600))
	content := container.New(layout.NewBorderLayout(nil, nil, left, nil), left, center)
	window.SetContent(content)

	// setup btns
	for _, obj := range devBtnList {
		obj.OnTapped = func() {
			DisplayDevice(obj.Text)
		}
	}
}

func ClearCenter() {
	if len(center.Objects) > 0 {
		center.Objects = []fyne.CanvasObject{}
	}
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
}

func DisplayBackup(bac string) {
	ClearCenter()
}
