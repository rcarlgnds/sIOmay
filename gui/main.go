package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"runtime"
	"sIOmay/gui/pages"
)

func main() {
	os := runtime.GOOS

	switch os {
	case "darwin":
		fmt.Println("Mac OS")
	case "linux":
		fmt.Println("Linux")
	case "windows":
		fmt.Println("windows")
	default:
		fmt.Println("Unknown OS")
	}

	application := app.New()

	window := application.NewWindow("sIOmay 🥟")
	window.Resize(fyne.NewSize(1100, 550))
	window.CenterOnScreen()
	window.SetContent(pages.Opening(window))

	window.ShowAndRun()
}
