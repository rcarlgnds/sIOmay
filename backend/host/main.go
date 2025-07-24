package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sIOmay/helpers"
	"strings"

	"github.com/go-vgo/robotgo"
)

// Helper function to check if error is due to connection being closed
func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		   strings.Contains(errStr, "connection reset") ||
		   strings.Contains(errStr, "broken pipe")
}

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
		
		// Set larger buffer and disable Nagle's algorithm for better performance
		connection.SetReadBuffer(65536)
		
		for {
			buffer := make([]byte, 1024)
			n, _, err := connection.ReadFromUDP(buffer)
			if err != nil {
				// Check if it's a "connection closed" error (normal during disconnect)
				if isConnectionClosed(err) {
					fmt.Println("Server connection closed - client shutting down gracefully")
				} else {
					fmt.Println("Error reading from server:", err)
				}
				break
			}
			
			data := buffer[:n]
			msg := string(data)
			
			// Check for disconnect command
			if msg == "DISCONNECT" {
				fmt.Println("Received disconnect command from server")
				break
			}
			
			if len(data) == 0 || data[0] != '{' {
				continue
			}
			
			var mouse helpers.Mouse
			err = json.Unmarshal(data, &mouse)
			if err != nil {
				continue
			}
			
			// Only move if position actually changed
			if mouse.Current.X != lastX || mouse.Current.Y != lastY {
				// Use instant movement - no smoothing
				robotgo.Move(mouse.Current.X, mouse.Current.Y)
				lastX, lastY = mouse.Current.X, mouse.Current.Y
			}
			
			// Handle clicks
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
		fmt.Println("Client disconnected and shutting down")
	}
