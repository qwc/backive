package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	app := app.New()
	w := app.NewWindow("Hello World!")

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}
