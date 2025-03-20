package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	helper "sIOmay/helpers"
	"time"
)

func AutoControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var selectedComputer []string

	connectButton := helper.InitConnectButton(&selectedComputer)
	connectButton.OnTapped = func() {
		window.SetContent(MinimizedAutoControlPanel(window))
	}

	backButton := func() {
		window.SetContent(Opening(window))
	}

	refreshButton := func() {
		window.SetContent(AutoControlPanel(window))
	}

	leftPart, computerBoxes := helper.InitLeftPanel(serverIP, &selectedComputer, connectButton)
	rightPart := helper.InitAutoRightPanel(window, serverIP, &selectedComputer, computerBoxes, connectButton, backButton, refreshButton)

	controlPanelPage := container.NewHSplit(leftPart, rightPart)
	controlPanelPage.SetOffset(0.6)

	return controlPanelPage
}

func MinimizedAutoControlPanel(window fyne.Window) fyne.CanvasObject {

	go func() {
		time.Sleep(time.Millisecond * 100)
		window.CenterOnScreen()
		window.RequestFocus()

	}()

	stopButton := widget.NewButton("Stop", func() {
		window.SetContent(AutoControlPanel(window))
	})

	return container.NewVBox(stopButton)
}
