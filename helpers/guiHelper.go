package helpers

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"sIOmay/object"
)

func InitHeader(window fyne.Window, serverIP string, backCallback func()) *fyne.Container {
	serverIPLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Server IP : %s", serverIP),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		backCallback()
	})

	return container.NewHBox(backButton, layout.NewSpacer(), serverIPLabel)
}

func InitConnectButton(selectedComputer *[]string) *widget.Button {
	return widget.NewButton("Connect", func() {
		fmt.Printf("Connecting to %v\n", selectedComputer)
		//window.Hide()
	})
}

func InitLeftPanel(serverIP string, selectedComputer *[]string, connectButton *widget.Button) (*fyne.Container, []fyne.CanvasObject) {
	computerBoxes, leftPart := UpdateComputerList(serverIP, selectedComputer, connectButton)
	return leftPart, computerBoxes
}

func InitManualRightPanel(
	window fyne.Window,
	serverIP string,
	selectedComputer *[]string,
	computerBoxes []fyne.CanvasObject,
	connectButton *widget.Button,
	backCallback func(),
	refreshCallback func(),
) *fyne.Container {
	header := InitHeader(window, serverIP, backCallback)

	networkAddressLabel := widget.NewLabel("Network Address")
	networkAddressInputField := widget.NewEntry()
	networkAddressInputField.SetPlaceHolder("Input Network Address Here")

	fromLabel := widget.NewLabel("From")
	fromInputField := widget.NewEntry()
	fromInputField.SetPlaceHolder("Input From Here")

	toLabel := widget.NewLabel("To")
	toInputField := widget.NewEntry()
	toInputField.SetPlaceHolder("Input To Here")

	middleContainer := container.NewVBox(networkAddressLabel, networkAddressInputField, fromLabel, fromInputField, toLabel, toInputField)

	selectAllCheckbox := widget.NewCheck("Select All", func(checked bool) {
		HandleSelectAll(checked, selectedComputer, computerBoxes, connectButton)
	})

	initiateConnectionButton := widget.NewButton("Initiate Connection", func() {

	})

	refreshButton := widget.NewButton("Refresh", func() {
		refreshCallback()
	})

	UpdateConnectButtonState(connectButton, *selectedComputer)

	return container.NewVBox(header, layout.NewSpacer(), middleContainer, layout.NewSpacer(), selectAllCheckbox, initiateConnectionButton, refreshButton, connectButton)
}

func InitAutoRightPanel(
	window fyne.Window,
	serverIP string,
	selectedComputer *[]string,
	computerBoxes []fyne.CanvasObject,
	connectButton *widget.Button,
	backCallback func(),
	refreshCallback func(),
) *fyne.Container {
	header := InitHeader(window, serverIP, backCallback)

	selectAllCheckbox := widget.NewCheck("Select All", func(checked bool) {
		HandleSelectAll(checked, selectedComputer, computerBoxes, connectButton)
	})

	refreshButton := widget.NewButton("Refresh", func() {
		fmt.Println("Refreshing Page")
		refreshCallback()
	})

	UpdateConnectButtonState(connectButton, *selectedComputer)

	return container.NewVBox(header, layout.NewSpacer(), selectAllCheckbox, refreshButton, connectButton)
}

func UpdateConnectButtonState(connectButton *widget.Button, selectedComputer []string) {
	if len(selectedComputer) > 0 {
		connectButton.Importance = widget.HighImportance
		connectButton.Enable()
	} else {
		connectButton.Disable()
	}
}

func HandleSelectAll(checked bool, selectedComputer *[]string, computerBoxes []fyne.CanvasObject, connectButton *widget.Button) {
	*selectedComputer = []string{} // Reset selectedComputer

	if checked {
		for _, button := range computerBoxes {
			if b, ok := button.(*widget.Button); ok && b.Importance != widget.WarningImportance {
				b.Importance = widget.SuccessImportance
				*selectedComputer = append(*selectedComputer, b.Text) // Tambahkan IP
				b.Refresh()
			}
		}
	} else {
		for _, button := range computerBoxes {
			if b, ok := button.(*widget.Button); ok && b.Importance != widget.WarningImportance {
				b.Importance = widget.MediumImportance
				b.Refresh()
			}
		}
	}

	UpdateConnectButtonState(connectButton, *selectedComputer)
}

func UpdateComputerList(serverIP string, selectedComputer *[]string, connectButton *widget.Button) ([]fyne.CanvasObject, *fyne.Container) {
	// Debug
	var computerList []object.Computer
	for i := 1; i <= 41; i++ {
		computer := object.Computer{
			IPAddress: fmt.Sprintf("10.22.65.%v", i),
			Status:    "Available",
		}
		computerList = append(computerList, computer)
	}

	//computerList := GetAllClients(serverIP)
	var computerBoxes []fyne.CanvasObject

	for _, computer := range computerList {
		button := widget.NewButton(computer.IPAddress, nil)

		switch computer.Status {
		case "Selected":
			button.Importance = widget.SuccessImportance
		case "Unavailable":
			button.Importance = widget.DangerImportance
		}

		if computer.IPAddress == serverIP {
			button.Importance = widget.WarningImportance
			button.OnTapped = func() {}
			button.Disable()
		}

		button.OnTapped = func(b *widget.Button) func() {
			return func() {
				if b.Importance != widget.SuccessImportance {
					b.Importance = widget.SuccessImportance
					*selectedComputer = append(*selectedComputer, serverIP)
				} else {
					b.Importance = widget.MediumImportance
					for i, ip := range *selectedComputer {
						if ip == b.Text {
							*selectedComputer = append((*selectedComputer)[:i], (*selectedComputer)[i+1:]...)
							break
						}
					}
				}
				UpdateConnectButtonState(connectButton, *selectedComputer)
				b.Refresh()
			}
		}(button)
		computerBoxes = append(computerBoxes, button)
	}
	return computerBoxes, container.NewGridWithColumns(5, computerBoxes...)
}
