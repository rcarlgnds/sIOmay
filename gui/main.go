package main

import (
	"flag"
	"fmt"
	"os"
	"sIOmay/gui/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	token := flag.String("token", "", "Token for authentication")
	flag.Parse()
	if *token == "" {
		fmt.Println("Usage: -token <your_token>")
		os.Exit(1)
	}
	application := app.New()
	window := application.NewWindow("sIOmay 🥟")
	window.Resize(fyne.NewSize(1100, 550))
	window.CenterOnScreen()
	window.SetContent(pages.Opening(window))

	window.ShowAndRun()
}
