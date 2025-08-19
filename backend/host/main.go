package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sIOmay/helpers"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

/*
#cgo LDFLAGS: mouse_move.o -luser32
extern void MoveMouse();
*/
import "C"

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		   strings.Contains(errStr, "connection reset") ||
		   strings.Contains(errStr, "broken pipe")
}
func smoothMove(fromX, fromY, toX, toY int) {
	dx := float64(toX - fromX)
	dy := float64(toY - fromY)
	distance := math.Sqrt(dx*dx + dy*dy)
	
	if distance < 3 {
		robotgo.Move(toX, toY)
		return
	}
	
	steps := int(math.Max(5, math.Min(25, distance/3)))
	
	for i := 1; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		
		progress = 1 - math.Pow(1-progress, 2.5)
		
		currentX := fromX + int(dx*progress)
		currentY := fromY + int(dy*progress)
		
		robotgo.Move(currentX, currentY)
		
		time.Sleep(time.Microsecond * 500)
	}
	
	robotgo.Move(toX, toY)
}
func processMouseMovement(x, y int, lastX, lastY *int, lastMoveTime *time.Time) {
	newX, newY := x, y
	
	if newX != *lastX || newY != *lastY {
		now := time.Now()
		if now.Sub(*lastMoveTime) >= time.Millisecond*16 {
			fmt.Printf("Moving mouse from (%d, %d) to (%d, %d)\n", *lastX, *lastY, newX, newY)
			
			smoothMove(*lastX, *lastY, newX, newY)
			*lastMoveTime = now
		} else {
			robotgo.Move(newX, newY)
		}
		
		*lastX, *lastY = newX, newY
	}
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
	var lastMoveTime time.Time
	var lastClickTime time.Time
	var lastClickButton int
	var lastClickCount int
	C.MoveMouse()
	connection.SetReadBuffer(65536)
	
	for {
			buffer := make([]byte, 1024)
			n, _, err := connection.ReadFromUDP(buffer)
			if err != nil {
				if isConnectionClosed(err) {
					fmt.Println("Server connection closed - client shutting down gracefully")
				} else {
					fmt.Println("Error reading from server:", err)
				}
				break
			}
			
			data := buffer[:n]
			msg := string(data)
			
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
		
		if mouse.Clicks > 0 {
			fmt.Printf("Received: X=%d, Y=%d, Button=%d, Clicks=%d, wasClicked=%t\n", 
				mouse.Current.X, mouse.Current.Y, mouse.Button, mouse.Clicks, wasClicked)
		} else if mouse.Clicks == 0 && (mouse.Button != 0 || wasClicked) {
			fmt.Printf("Received RELEASE: X=%d, Y=%d, Button=%d, Clicks=%d, wasClicked=%t\n", 
				mouse.Current.X, mouse.Current.Y, mouse.Button, mouse.Clicks, wasClicked)
		}
		
		processMouseMovement(mouse.Current.X, mouse.Current.Y, &lastX, &lastY, &lastMoveTime)
		
		now := time.Now()
		clickDuplicate := false
		
		if mouse.Clicks > 0 && mouse.Button == lastClickButton && 
		   mouse.Clicks == lastClickCount && 
		   now.Sub(lastClickTime) < time.Millisecond*100 {
			clickDuplicate = true
			fmt.Printf("Duplicate click detected - ignoring (Button %d, Clicks %d, TimeSince: %v)\n", 
				mouse.Button, mouse.Clicks, now.Sub(lastClickTime))
		}
		
		if mouse.Clicks > 0 && !wasClicked && !clickDuplicate {
			fmt.Printf("Click detected: Button %d, Clicks %d\n", mouse.Button, mouse.Clicks)
			if mouse.Button == 1 {
				robotgo.MouseClick("left", false)
				fmt.Println("Executed left click")
			} else if mouse.Button == 2 {
				robotgo.MouseClick("right", false)
				fmt.Println("Executed right click")
			}
			wasClicked = true
			lastClickTime = now
			lastClickButton = mouse.Button
			lastClickCount = mouse.Clicks
		} else if mouse.Clicks == 0 {
			if wasClicked {
				fmt.Println("Click released - resetting wasClicked")
			}
			wasClicked = false
		}
		}
		fmt.Println("Client disconnected and shutting down")
	}
