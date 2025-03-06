package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Opening(window fyne.Window) fyne.CanvasObject {
	autoButton := widget.NewButtonWithIcon("  Auto Mode  ", theme.DesktopIcon(), func() {
		window.SetContent(AutoControlPanel(window))
	})
	manualButton := widget.NewButtonWithIcon("  Manual Mode  ", theme.BrokenImageIcon(), func() {
		window.SetContent(ManualControlPanel(window))
	})

	autoButton.Resize(fyne.NewSize(200, 400))
	center := container.NewHSplit(autoButton, manualButton)

	return container.NewCenter(center)
}
