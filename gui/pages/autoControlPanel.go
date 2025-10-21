package pages

import (
	"fmt"
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
	var computerBoxes []fyne.CanvasObject

	connectButton := controller.InitConnectButtonWithWindow(&selectedComputer, window)

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

	// Callback to clear selections and update UI when connections change
	clearSelections := func() {
		// Clear all selections from the computer boxes
		for _, computerBox := range computerBoxes {
			if btn, ok := computerBox.(*widget.Button); ok && btn.Importance == widget.SuccessImportance {
				btn.Importance = widget.MediumImportance
				btn.Refresh()
			}
		}
		selectedComputer = []string{}
		updateButtonState()
	}

	leftPart, computerBoxes := helper.InitLeftPanelWithConnectionInfo(serverIP, &selectedComputer, connectButton, updateButtonState)
	rightPart, connectedListLabel := helper.InitAutoRightPanelWithConnectionInfo(window, serverIP, &selectedComputer, computerBoxes, connectButton, backButton, refreshButton, updateButtonState)

	// Start a goroutine to periodically update the UI and check for connection changes
	go func() {
		lastConnectedCount := len(controller.GetConnectedClients())
		for {
			time.Sleep(500 * time.Millisecond) // Update every 500ms

			connectedClients := controller.GetConnectedClients()
			currentConnectedCount := len(connectedClients)

			// Update connected computers display
			if currentConnectedCount == 0 {
				connectedListLabel.SetText("None")
			} else {
				connectedText := fmt.Sprintf("%d connected: %v", currentConnectedCount, connectedClients)
				if len(connectedText) > 50 {
					connectedText = fmt.Sprintf("%d connected", currentConnectedCount)
				}
				connectedListLabel.SetText(connectedText)
			}

			if currentConnectedCount != lastConnectedCount {
				// Connection state changed, clear selections
				clearSelections()
				lastConnectedCount = currentConnectedCount
			}
			updateButtonState()
		}
	}()

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
