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
		// Define and parse the `-from` flag
		fromIP := flag.String("from", "", "IP address of the controller (e.g., 10.22.65.133:8080)")
		flag.Parse()

		if *fromIP == "" {
			fmt.Println("Error: Please specify the controller IP using the -from flag")
			fmt.Println("Usage: client.exe -from 10.22.65.133:8080")
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

		buffer := make([]byte, 1024)
		var lastX, lastY int

		for {
			n, _, err := connection.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				break
			}
		
			data := buffer[:n]
			msg := string(data)
		
			// Ignore any non-JSON messages
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
		
			// Only act on valid JSON messages
			mouseData := mouseMessage
			if mouseData.Current.X != lastX || mouseData.Current.Y != lastY {
				robotgo.Move(mouseData.Current.X, mouseData.Current.Y)
				lastX, lastY = mouseData.Current.X, mouseData.Current.Y
			}

			// Handle Mouse Wheel
			//if mouseData.Actions.Wheel != nil {
			//	if mouseData.Actions.Wheel.Event == "scroll_up" {
			//		robotgo.ScrollDir(mouseData.Actions.Wheel.Rotation, "up")
			//	} else if mouseData.Actions.Wheel.Event == "scroll_down" {
			//		robotgo.ScrollDir(mouseData.Actions.Wheel.Rotation, "down")
			//	}
			//}

			// Mouse Action Handling
			//mouseAction := strings.TrimSpace(coordinates[2])
			//switch mouseAction {
			//case "left_click":
			//	fmt.Println("Performing left click")
			//	robotgo.Click("left")
			//case "right_click":
			//	fmt.Println("Performing right click")
			//	robotgo.Click("right")
			//case "middle_click":
			//	fmt.Println("Performing middle click")
			//	robotgo.Click("center")
			//case "double_click":
			//	fmt.Println("Performing double left click")
			//	robotgo.Click("left", true)
			//case "move_click":
			//	fmt.Println("Moving and clicking at:", x, y)
			//	robotgo.MoveClick(x, y)
			//case "move_smooth_click":
			//	fmt.Println("Smoothly moving and clicking at:", x, y)
			//	robotgo.MoveSmooth(x, y)
			//	robotgo.Click("left")
			//case "drag_start":
			//	fmt.Println("Starting drag operation")
			//	robotgo.MouseDown("left")
			//case "drag_end":
			//	fmt.Println("Ending drag operation")
			//	robotgo.MouseUp("left")
			//case "scroll_up":
			//	fmt.Println("Scrolling up")
			//	robotgo.ScrollDir(10, "up")
			//case "scroll_down":
			//	fmt.Println("Scrolling down")
			//	robotgo.ScrollDir(10, "down")
			//case "scroll_left":
			//	fmt.Println("Scrolling left")
			//	robotgo.ScrollDir(10, "left")
			//case "scroll_right":
			//	fmt.Println("Scrolling right")
			//	robotgo.ScrollDir(10, "right")
			//case "scroll_smooth":
			//	fmt.Println("Performing smooth scroll")
			//	robotgo.ScrollSmooth(-10)
			//case "hold_left":
			//	fmt.Println("Holding left mouse button")
			//	robotgo.MouseDown("left")
			//case "release_left":
			//	fmt.Println("Releasing left mouse button")
			//	robotgo.MouseUp("left")
			//case "hold_right":
			//	fmt.Println("Holding right mouse button")
			//	robotgo.MouseDown("right")
			//case "release_right":
			//	fmt.Println("Releasing right mouse button")
			//	robotgo.MouseUp("right")
			//default:
			//	fmt.Println("Unknown mouse action:", mouseAction)
			//}

			// Process Keyboard Data
			//keyboardData := message.KeyboardData
			//fmt.Println("Host Keyboard:", keyboardData)
			//if keyboardData.Press {
			//	fmt.Println("Key Pressed:", keyboardData.Key)
			//	// Simulate key press
			//	robotgo.KeyTap(keyboardData.Key)
			//}
		}
	}
