package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sIOmay/helpers"

	"github.com/go-vgo/robotgo"
)

	func main() {
		fromIP := flag.String("from", "", "IP address of the controller (e.g., 10.22.65.133:8080)")
		flag.Parse()
		if *fromIP == "" {
			os.Exit(1)
		}
		serverAddress, err := net.ResolveUDPAddr("udp4", *fromIP)
		if err != nil {
			fmt.Println("Error resolving server address:", err)
			return
		}
		connection, err := net.DialUDP("udp4", nil, serverAddress)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer connection.Close()
		_, err = connection.Write([]byte("Client ready"))
		if err != nil {
			fmt.Println("Error sending registration message:", err)
			return
		}
		var lastX, lastY int
		var wasClicked bool
		for {
			buffer := make([]byte,256)
			n, _, err := connection.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				break
			}
			data := buffer[:n]
			msg := string(data)
			if len(data) == 0 || data[0] != '{' {
				fmt.Println("Non-JSON server message:", msg)
				continue
			}
		
			var mouseMessage helpers.Mouse
			err = json.Unmarshal(data, &mouseMessage)
			if err != nil {
				fmt.Println("Invalid JSON from server:", err)
				continue
			}
		
			mouse := mouseMessage

			if mouse.Current.X != lastX || mouse.Current.Y != lastY {
				robotgo.Move(mouse.Current.X, mouse.Current.Y)
				lastX, lastY = mouse.Current.X, mouse.Current.Y
			}
			if mouse.Clicks > 0 && !wasClicked {
		if mouse.Button == 1 {
			robotgo.MouseClick("left", false)
		} else if mouse.Button == 2 {
			robotgo.MouseClick("right", false)
		}
		wasClicked = true
	} else if mouse.Clicks == 0 {
		wasClicked = false
	}
		}
	}
