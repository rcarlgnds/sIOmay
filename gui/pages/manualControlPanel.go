package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	helper "sIOmay/helpers"
)

func ManualControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var selectedComputer []string

	connectButton := helper.InitConnectButton(&selectedComputer)

	backButton := func() {
		window.SetContent(Opening(window))
	}

	refreshButton := func() {
		window.SetContent(ManualControlPanel(window))
	}

	leftPart, computerBoxes := helper.InitLeftPanel(serverIP, &selectedComputer, connectButton)
	rightPart := helper.InitManualRightPanel(window, serverIP, &selectedComputer, computerBoxes, connectButton, backButton, refreshButton)

	controlPanelPage := container.NewHSplit(leftPart, rightPart)
	controlPanelPage.SetOffset(0.6)

	return controlPanelPage
}
