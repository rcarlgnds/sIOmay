package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Opening(window fyne.Window) fyne.CanvasObject {
	startButton := widget.NewButton("Start", func() {
		window.SetContent(ControlPanel(window))
	})

	return container.NewCenter(startButton)
}
