package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"sIOmay/gui/pages"
)

func main() {
	application := app.New()

	window := application.NewWindow("sIOmay 🥟")
	window.Resize(fyne.NewSize(800, 500))

	window.CenterOnScreen()
	window.SetContent(pages.Opening(window))

	window.ShowAndRun()
}
