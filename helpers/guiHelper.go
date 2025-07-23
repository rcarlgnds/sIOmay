package helpers

import (
	"fmt"
	"log"
	"runtime"
	"sIOmay/object"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-ping/ping"
)

func InitHeader(window fyne.Window, serverIP string, backCallback func()) *fyne.Container {
	os := runtime.GOOS

	var osType string

	switch os {
	case "darwin":
		osType = "Mac OS"
	case "linux":
		osType = "Linux"
	case "windows":
		osType = "Windows"
	default:
		osType = "Unknown OS"
	}

	backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		backCallback()
	})

	serverIPLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("| %s |", serverIP),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	osTypeLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("| %s |", osType),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return container.NewHBox(backButton, layout.NewSpacer(), serverIPLabel, osTypeLabel)
}

func InitLeftPanel(serverIP string, selectedComputer *[]string, connectButton *widget.Button) (*fyne.Container, []fyne.CanvasObject) {
	computerBoxes, leftPart := UpdateComputerList(serverIP, selectedComputer, connectButton)
	return leftPart, computerBoxes
}

func InitManualRightPanel(
	window fyne.Window,
	serverIP string,
	backCallback func(),
	refreshCallback func(),
) (*fyne.Container, *widget.Entry, *widget.Entry, *widget.Entry, *widget.Button) {
	header := InitHeader(window, serverIP, backCallback)

	networkAddressLabel := widget.NewLabel("Network Address (e.g., 192.168.1)")
	networkAddressInputField := widget.NewEntry()
	networkAddressInputField.SetPlaceHolder("192.168.1")

	fromLabel := widget.NewLabel("From")
	fromInputField := widget.NewEntry()
	fromInputField.SetPlaceHolder("101")

	toLabel := widget.NewLabel("To")
	toInputField := widget.NewEntry()
	toInputField.SetPlaceHolder("141")

	scanButton := widget.NewButton("Scan IP Range", nil) 
	scanButton.Importance = widget.HighImportance

	refreshButton := widget.NewButton("Refresh Page", func() {
		refreshCallback()
	})
	rightPanel := container.NewVBox(
		header,
		layout.NewSpacer(),
		networkAddressLabel,
		networkAddressInputField,
		fromLabel,
		fromInputField,
		toLabel,
		toInputField,
		layout.NewSpacer(),
		scanButton,
		refreshButton,
	)
	return rightPanel, networkAddressInputField, fromInputField, toInputField, scanButton
}

func GenerateComputerGrid(
	computerList []object.Computer,
	serverIP string,
	selectedComputer *[]string,
	connectButton *widget.Button,
) (*fyne.Container, []fyne.CanvasObject) {

	var computerBoxes []fyne.CanvasObject

	for _, computer := range computerList {
		ipAddr := computer.IPAddress
		button := widget.NewButton(ipAddr, nil)

		if ipAddr == serverIP {
			button.Importance = widget.WarningImportance
			button.Disable()
		} else {
			button.Importance = widget.LowImportance
			button.Disable()
			go pingAndUpdate(button, ipAddr)
		}

		button.OnTapped = func() {
			isCurrentlySelected := false
			for _, ip := range *selectedComputer {
				if ip == button.Text {
					isCurrentlySelected = true
					break
				}
			}

			if !isCurrentlySelected {
				button.Importance = widget.SuccessImportance
				*selectedComputer = append(*selectedComputer, button.Text)
			} else {
				button.Importance = widget.MediumImportance
				newSelection := []string{}
				for _, ip := range *selectedComputer {
					if ip != button.Text {
						newSelection = append(newSelection, ip)
					}
				}
				*selectedComputer = newSelection
			}
			UpdateConnectButtonState(connectButton, *selectedComputer)
			button.Refresh()
		}
		computerBoxes = append(computerBoxes, button)
	}
	return container.NewGridWithColumns(5, computerBoxes...), computerBoxes
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
		connectButton.Importance = widget.MediumImportance
		connectButton.Disable()
	}
}

func HandleSelectAll(checked bool, selectedComputer *[]string, computerBoxes []fyne.CanvasObject, connectButton *widget.Button) {
	*selectedComputer = []string{}

	if checked {
		for _, item := range computerBoxes {
			if button, ok := item.(*widget.Button); ok && !button.Disabled() && button.Importance != widget.WarningImportance {
				button.Importance = widget.SuccessImportance
				*selectedComputer = append(*selectedComputer, button.Text)
				button.Refresh()
			}
		}
	} else {
		for _, item := range computerBoxes {
			if button, ok := item.(*widget.Button); ok && !button.Disabled() && button.Importance != widget.WarningImportance {
				button.Importance = widget.MediumImportance
				button.Refresh()
			}
		}
	}

	UpdateConnectButtonState(connectButton, *selectedComputer)
}

func pingAndUpdate(button *widget.Button, ipAddress string) {
	pinger, err := ping.NewPinger(ipAddress)
	if err != nil {
		log.Printf("Failed to create pinger for %s: %v", ipAddress, err)
		return
	}

	pinger.Count = 1
	pinger.Timeout = time.Second * 1
	pinger.SetPrivileged(true) 

	err = pinger.Run() 
	if err != nil {
		log.Printf("Pinger failed to run for %s: %v", ipAddress, err)
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		
		button.Enable()
		button.Importance = widget.MediumImportance
	} else {
		
		button.Importance = widget.DangerImportance
		button.Disable()
	}
	button.Refresh()
}

func UpdateComputerList(serverIP string, selectedComputer *[]string, connectButton *widget.Button) ([]fyne.CanvasObject, *fyne.Container) {
	var computerList []object.Computer

	

	
	lastDotIndex := strings.LastIndex(serverIP, ".")
	networkPrefix := "10.22.65" 

	if lastDotIndex != -1 {
		
		
		networkPrefix = serverIP[:lastDotIndex]
	} else {
		log.Printf("Warning: Could not determine network prefix from server IP '%s'. Using default.", serverIP)
	}


	
	for i := 101; i <= 141; i++ {
		computer := object.Computer{
			IPAddress: fmt.Sprintf("%s.%v", networkPrefix, i),
			Status:    "Available",
		}
		computerList = append(computerList, computer)
	}
	
	


	
	var computerBoxes []fyne.CanvasObject

	for _, computer := range computerList {
		ipAddr := computer.IPAddress
		button := widget.NewButton(ipAddr, nil)

		if ipAddr == serverIP {
			button.Importance = widget.WarningImportance
			button.Disable()
		} else {
			button.Importance = widget.LowImportance
			button.Disable()
			go pingAndUpdate(button, ipAddr)
		}

		button.OnTapped = func() {
			isCurrentlySelected := false
			for _, ip := range *selectedComputer {
				if ip == button.Text {
					isCurrentlySelected = true
					break
				}
			}

			if !isCurrentlySelected {
				button.Importance = widget.SuccessImportance
				*selectedComputer = append(*selectedComputer, button.Text)
			} else {
				button.Importance = widget.MediumImportance
				newSelection := []string{}
				for _, ip := range *selectedComputer {
					if ip != button.Text {
						newSelection = append(newSelection, ip)
					}
				}
				*selectedComputer = newSelection
			}
			UpdateConnectButtonState(connectButton, *selectedComputer)
			button.Refresh()
		}
		computerBoxes = append(computerBoxes, button)
	}
	return computerBoxes, container.NewGridWithColumns(5, computerBoxes...)
}