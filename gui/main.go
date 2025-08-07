package main

import (
	"sIOmay/gui/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	application := app.New()

	window := application.NewWindow("sIOmay 🥟")
	window.Resize(fyne.NewSize(1100, 550))
	window.CenterOnScreen()
	window.SetContent(pages.Opening(window))

	window.ShowAndRun()
}
