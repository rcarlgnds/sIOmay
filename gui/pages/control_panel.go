package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

func ControlPanel(window fyne.Window) fyne.CanvasObject {
	// Header
	serverIPLabel := widget.NewLabel("Server IP : 10.22.225.141")
	backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		window.SetContent(Opening(window))
	})

	header := container.NewHBox(backButton, layout.NewSpacer(), serverIPLabel)

	// Body
	var computerList []string
	for i := 1; i <= 40; i++ {
		computerList = append(computerList, fmt.Sprintf("PC %02d", i))
	}

	var computerBoxes []fyne.CanvasObject
	for _, name := range computerList {
		rect := canvas.NewRectangle(color.White)
		rect.SetMinSize(fyne.NewSize(50, 20))

		text := canvas.NewText(name, color.Black)
		text.Alignment = fyne.TextAlignCenter
		text.TextSize = 14

		box := container.NewStack(rect, container.New(layout.NewCenterLayout(), text))
		computerBoxes = append(computerBoxes, box)
	}
	computerGrid := container.NewGridWithColumns(5, computerBoxes...)

	// Footer
	scanButton := widget.NewButton("Scan", func() {
		fmt.Println("Scanning for computers...")
	})

	connectButton := widget.NewButton("Connect", func() {
		fmt.Println("Connecting...")
	})

	footer := container.NewHBox(scanButton, layout.NewSpacer(), connectButton)

	controlPanelPage := container.NewBorder(header, footer, nil, nil, computerGrid)
	return controlPanelPage
}
