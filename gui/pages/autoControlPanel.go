package pages

import (
	"sIOmay/controller"
	helper "sIOmay/helpers"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func AutoControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var selectedComputer []string

	connectButton := controller.InitConnectButton(&selectedComputer)

	backButton := func() {
		window.SetContent(Opening(window))
	}

	refreshButton := func() {
		window.SetContent(AutoControlPanel(window))
	}

	// Create a periodic refresh function to update button state
	updateButtonState := func() {
		isConnected := controller.IsConnected()
		connectedClients := controller.GetConnectedClients()
		helper.UpdateConnectButtonStateWithConnectionInfo(connectButton, selectedComputer, isConnected, connectedClients)
	}

	leftPart, computerBoxes := helper.InitLeftPanelWithConnectionInfo(serverIP, &selectedComputer, connectButton, updateButtonState)
	rightPart := helper.InitAutoRightPanelWithConnectionInfo(window, serverIP, &selectedComputer, computerBoxes, connectButton, backButton, refreshButton, updateButtonState)

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
