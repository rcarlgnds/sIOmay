package main

import (
	"flag"
	"fmt"
	"os"
	"sIOmay/controller"
	"sIOmay/gui/pages"
	"sIOmay/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Application crashed with panic: %v\n", r)
			fmt.Println("Performing emergency client disconnection...")
			controller.DisconnectFromClients()
		}
	}()

	token := flag.String("token", "", "Token for authentication")
	flag.Parse()
	if *token == "" {
		fmt.Println("Usage: -token <your_token>")
		os.Exit(1)
	}
	controller.SetAuthToken(*token)
	_, err := controller.VerifyToken(*token)
	if err != nil {
		fmt.Println("Error verifying token:", err)
		os.Exit(1)
	}else{
		fmt.Println("DEXTER BODOH")
	}

	application := app.New()
	window := application.NewWindow("sIOmay 🥟")
	window.Resize(fyne.NewSize(1100, 550))
	window.CenterOnScreen()

	controller.SetGlobalWindow(window)

	helpers.StartGlobalHotkeyListener(controller.ShowWindow, controller.DisconnectFromClients)

	window.SetCloseIntercept(func() {
		fmt.Println("Application closing. Disconnecting all clients...")
		controller.DisconnectFromClients()
		window.Close()
	})

	window.SetContent(pages.Opening(window))

	window.ShowAndRun()
}
