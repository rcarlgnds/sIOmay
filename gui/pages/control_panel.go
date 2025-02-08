package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	helper "sIOmay/helpers"
)

func ControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var selectedComputer []string

	connectButton := helper.InitConnectButton(&selectedComputer)

	leftPart, computerBoxes := helper.InitLeftPanel(serverIP, &selectedComputer, connectButton)

	rightPart := helper.InitRightPanel(window, serverIP, &selectedComputer, computerBoxes, connectButton)

	controlPanelPage := container.NewHSplit(leftPart, rightPart)
	controlPanelPage.SetOffset(0.6)

	return controlPanelPage
}
