package pcbox

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type Box struct {
	IP     string
	Status string
}

func ComputerBox(computer *Box, onClick func()) fyne.CanvasObject {
	var boxColor color.Color
	var _ color.Color

	switch computer.Status {
	case "Default":
		boxColor = color.White
		_ = color.Black
	case "Selected":
		boxColor = color.White
		_ = color.Black
	case "Connected":
		boxColor = color.RGBA{0, 255, 0, 255}
		_ = color.White
	}

	// Background
	rect := canvas.NewRectangle(boxColor)
	rect.SetMinSize(fyne.NewSize(100, 80))

	// Button
	button := widget.NewButton(computer.IP, onClick)
	button.Importance = widget.HighImportance // Memberi efek lebih jelas

	// Stack container
	box := container.NewStack(rect, container.NewCenter(button))

	return box
}
