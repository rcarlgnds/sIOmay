package pcbox

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"image/color"
)

type Computer struct {
	Name   string
	IP     string
	Status string
}

func ComputerBox(computer *Computer, onClick func()) fyne.CanvasObject {
	var boxColor color.Color
	var textColor color.Color
	var borderColor color.Color
	var borderWidth int

	switch computer.Status {
	case "Default":
		boxColor = color.White
		textColor = color.Black
		borderColor = color.Transparent
		borderWidth = 0
	case "Selected":
		boxColor = color.White
		textColor = color.Black
		borderColor = color.RGBA{0, 255, 0, 255}
		borderWidth = 2
	case "Connected":
		boxColor = color.RGBA{0, 255, 0, 255}
		textColor = color.White
		borderColor = color.RGBA{0, 255, 0, 255}
		borderWidth = 2
	}

	// Box
	rect := canvas.NewRectangle(boxColor)
	rect.SetMinSize(fyne.NewSize(100, 80))
	rect.StrokeColor = borderColor
	rect.StrokeWidth = float32(borderWidth)

	// Text
	text := canvas.NewText(computer.Name, textColor)
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 14

	box := container.NewStack(rect, container.New(layout.NewCenterLayout(), text))

	// Add a click listener to change the state when the user clicks a box
	//box.OnTapped = onClick

	return box
}
