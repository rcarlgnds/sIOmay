package pages

import (
	"fmt"
	"log"
	"os/exec"
	"sIOmay/helpers"
	"sIOmay/object"
	"strconv"

	controller "sIOmay/controller"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var runningServerCmd *exec.Cmd

func ManualControlPanel(window fyne.Window) fyne.CanvasObject {
	serverIP, _ := helpers.GetServerIP()
	var selectedComputer []string
	var computerBoxes []fyne.CanvasObject


	leftContent := container.NewCenter(widget.NewLabel("Enter an IP range and click Scan..."))
	leftScroll := container.NewScroll(leftContent)
	
	connectButton := controller.InitConnectButton(&selectedComputer)

	backButton := func() {
		window.SetContent(Opening(window))
	}
	refreshButton := func() {
		window.SetContent(ManualControlPanel(window))
	}

	rightPart, networkAddressInput, fromInput, toInput, scanButton := helpers.InitManualRightPanel(window, serverIP, backButton, refreshButton)

	scanButton.OnTapped = func() {
		prefix := networkAddressInput.Text
		fromStr := fromInput.Text
		toStr := toInput.Text

		if prefix == "" || fromStr == "" || toStr == "" {
			log.Println("Error: All fields must be filled.")
			return
		}
		from, err1 := strconv.Atoi(fromStr)
		to, err2 := strconv.Atoi(toStr)

		if err1 != nil || err2 != nil || from > to {
			log.Println("Error: Invalid 'From' or 'To' values.")
			return
		}

		var computerList []object.Computer
		for i := from; i <= to; i++ {
			computer := object.Computer{
				IPAddress: fmt.Sprintf("%s.%d", prefix, i),
				Status:    "Available",
			}
			computerList = append(computerList, computer)
		}

		newGrid, newBoxes := helpers.GenerateComputerGrid(computerList, serverIP, &selectedComputer, connectButton)

		leftScroll.Content = newGrid
		leftScroll.Refresh()
		computerBoxes = newBoxes 
	}

	selectAllCheckbox := widget.NewCheck("Select All", func(checked bool) {
		helpers.HandleSelectAll(checked, &selectedComputer, computerBoxes, connectButton)
	})

	rightPart.Add(widget.NewSeparator())
	rightPart.Add(selectAllCheckbox)
	rightPart.Add(connectButton) 

	controlPanelPage := container.NewHSplit(leftScroll, rightPart)
	controlPanelPage.SetOffset(0.6)

	return controlPanelPage
}