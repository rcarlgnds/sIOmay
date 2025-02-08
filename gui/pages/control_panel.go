package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	helper "sIOmay/helpers"
	"sIOmay/object"
)

func updateConnectButtonState(connectButton *widget.Button, selectedComputer []string) {
	if len(selectedComputer) > 0 {
		connectButton.Importance = widget.HighImportance
		connectButton.Enable()
	} else {
		connectButton.Disable()
	}
}

func ControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helper.GetServerIP()
	var connectButton *widget.Button

	// Left Part
	//computerList := helper.GetAllClients(serverIP)
	// DEBUG
	var computerList []object.Computer
	for i := 1; i <= 41; i++ {
		computerList = append(computerList, object.Computer{
			ComputerIP: fmt.Sprintf("%s.%d", helper.GetNetworkPrefix(serverIP), i),
			Status:     "Available",
		})
	}

	var computerBoxes []fyne.CanvasObject
	var selectedComputer []string
	for _, computer := range computerList {
		button := widget.NewButton(computer.ComputerIP, nil)

		switch computer.Status {
		case "Selected":
			button.Importance = widget.SuccessImportance
		case "Unavailable":
			button.Importance = widget.DangerImportance
		}

		if computer.ComputerIP == serverIP {
			button.Importance = widget.WarningImportance
			button.OnTapped = func() {}
			button.Disable()
		}

		button.OnTapped = func(b *widget.Button) func() {
			return func() {
				if b.Importance != widget.SuccessImportance {
					b.Importance = widget.SuccessImportance
					selectedComputer = append(selectedComputer, b.Text)
				} else {
					b.Importance = widget.MediumImportance
					for i, ip := range selectedComputer {
						if ip == b.Text {
							selectedComputer = append(selectedComputer[:i], selectedComputer[i+1:]...)
							break
						}
					}
				}
				updateConnectButtonState(connectButton, selectedComputer)
				b.Refresh()
			}
		}(button)
		button.Resize(fyne.NewSize(0, 0))
		computerBoxes = append(computerBoxes, button)
	}
	leftPart := container.NewGridWithColumns(5, computerBoxes...)

	// Right Part
	// [Upper Right]
	serverIPLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Server IP : %s", serverIP),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true, Italic: false},
	)

	backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		window.SetContent(Opening(window))
	})
	header := container.NewHBox(backButton, layout.NewSpacer(), serverIPLabel)

	selectAllCheckbox := widget.NewCheck("Select All", func(checked bool) {
		if checked {
			// Select all PCs
			for _, button := range computerBoxes {
				if b, ok := button.(*widget.Button); ok && b.Importance != widget.WarningImportance {
					b.Importance = widget.SuccessImportance
					selectedComputer = append(selectedComputer, b.Text)
					b.Refresh()
				}
			}
		} else {
			// Unselect all PCs
			selectedComputer = nil
			for _, button := range computerBoxes {
				if b, ok := button.(*widget.Button); ok && b.Importance != widget.WarningImportance {
					b.Importance = widget.MediumImportance
					b.Refresh()
				}
			}
		}
		updateConnectButtonState(connectButton, selectedComputer)
	})

	refreshButton := widget.NewButton("Refresh", func() {
		fmt.Println("Scanning for computers...")
	})

	connectButton = widget.NewButton("Connect", func() {
		fmt.Printf("Connecting to %v\n", selectedComputer)
		//window.Hide()
	})

	updateConnectButtonState(connectButton, selectedComputer)
	rightPart := container.NewVBox(header, layout.NewSpacer(), selectAllCheckbox, refreshButton, connectButton)

	controlPanelPage := container.NewHSplit(leftPart, rightPart)
	controlPanelPage.SetOffset(0.6)

	return controlPanelPage
}
