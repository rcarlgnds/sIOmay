package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	helper "sIOmay/helpers"
)

func AutoControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var selectedComputer []string

	connectButton := helper.InitConnectButton(&selectedComputer)

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
