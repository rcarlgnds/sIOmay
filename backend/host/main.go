package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sIOmay/core"
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
	
	if distance < 5 {
		robotgo.Move(toX, toY)
		return
	}
	
	steps := int(math.Max(8, math.Min(30, distance/5)))
	
	for i := 1; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		
		progress = 1 - math.Pow(1-progress, 3)
		
		currentX := fromX + int(dx*progress)
		currentY := fromY + int(dy*progress)
		
		robotgo.Move(currentX, currentY)
		
		time.Sleep(time.Microsecond * 1)
	}
}

func main() {
	fromIP := flag.String("from", "", "IP address of the controller (e.g., 10.22.65.133:8080)")
	flag.Parse()
	if *fromIP == "" {
		os.Exit(1)
	}

	connection, err := setupConnection(*fromIP)
	if err != nil {
		return
	}
	defer connection.Close()

	startMouseControl(connection)
}

func setupConnection(fromIP string) (*net.UDPConn, error) {
	serverAddress, err := net.ResolveUDPAddr("udp4", fromIP)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return nil, err
	}

	connection, err := net.DialUDP("udp4", nil, serverAddress)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil, err
	}

	_, err = connection.Write([]byte("Client ready"))
	if err != nil {
		fmt.Println("Error sending registration message:", err)
		return nil, err
	}

	return connection, nil
}

func startMouseControl(connection *net.UDPConn) {
	C.MoveMouse()
	var lastX, lastY int
	var lastLeftClick, lastRightClick, lastMiddleClick bool
	var lastMoveTime time.Time

	fmt.Println("Starting byte-based mouse control with smooth interpolation...")
	
	for {
		if !processMouseData(connection, &lastX, &lastY, &lastLeftClick, &lastRightClick, &lastMiddleClick, &lastMoveTime) {
			break
		}
	}
	fmt.Println("Client disconnected and shutting down")
}

func processMouseData(connection *net.UDPConn, lastX, lastY *int, lastLeftClick, lastRightClick, lastMiddleClick *bool, lastMoveTime *time.Time) bool {
	buffer := make([]byte, 1024)
	n, _, err := connection.ReadFromUDP(buffer)
	if err != nil {
		if isConnectionClosed(err) {
			fmt.Println("Server connection closed - client shutting down gracefully")
		} else {
			fmt.Println("Error reading from server:", err)
		}
		return false
	}

	data := buffer[:n]

	if n > 0 && string(data) == "DISCONNECT" {
		fmt.Println("Received disconnect command from server")
		return false
	}

	// Process byte data (expecting 7 bytes)
	if n == 7 {
		executeMouseActions(data, lastX, lastY, lastLeftClick, lastRightClick, lastMiddleClick, lastMoveTime)
	} else if n > 0 {
		fmt.Printf("Received unexpected data size: %d bytes (expected 7)\n", n)
	}

	return true
}

func executeMouseActions(data []byte, lastX, lastY *int, lastLeftClick, lastRightClick, lastMiddleClick *bool, lastMoveTime *time.Time) {
	// Create Bytedata from received bytes
	byteData := core.NewBytedata()
	// Copy received data to the internal byte array
	for i := 0; i < 7; i++ {
		byteData.Bytes()[i] = data[i]
	}

	// Optional: Print received data for debugging (comment out for production)
	// fmt.Printf("=== RECEIVED MOUSE DATA ===\n")
	// fmt.Printf("Raw bytes: %v\n", data)
	// fmt.Printf("Hex: %X\n", data)

	processMouseMovement(byteData, lastX, lastY, lastMoveTime)
	processMouseClicks(byteData, lastLeftClick, lastRightClick, lastMiddleClick)
	processMouseScroll(byteData)

	// fmt.Println("=============================")
}

func processMouseMovement(byteData *core.Bytedata, lastX, lastY *int, lastMoveTime *time.Time) {
	if byteData.HasMouseMove() {
		x, y := byteData.GetMousePosition()
		newX, newY := int(x), int(y)
		
		// Rate limiting: don't move more than 120 times per second
		now := time.Now()
		if now.Sub(*lastMoveTime) < time.Millisecond*8 {
			return
		}
		*lastMoveTime = now
		
		// Only move if position actually changed
		if newX != *lastX || newY != *lastY {
			fmt.Printf("Moving mouse from (%d, %d) to (%d, %d)\n", *lastX, *lastY, newX, newY)
			
			// Use smooth interpolated movement instead of instant jump
			smoothMove(*lastX, *lastY, newX, newY)
			
			*lastX, *lastY = newX, newY
		}
	}
}

func processMouseClicks(byteData *core.Bytedata, lastLeftClick, lastRightClick, lastMiddleClick *bool) {
	if byteData.HasMouseClickLeft() && !*lastLeftClick {
		fmt.Println("Executing left click")
		robotgo.MouseClick("left", false)
		*lastLeftClick = true
	} else if !byteData.HasMouseClickLeft() {
		*lastLeftClick = false
	}

	if byteData.HasMouseClickRight() && !*lastRightClick {
		fmt.Println("Executing right click")
		robotgo.MouseClick("right", false)
		*lastRightClick = true
	} else if !byteData.HasMouseClickRight() {
		*lastRightClick = false
	}

	if byteData.HasMouseMiddleClick() && !*lastMiddleClick {
		fmt.Println("Executing middle click")
		robotgo.MouseClick("center", false)
		*lastMiddleClick = true
	} else if !byteData.HasMouseMiddleClick() {
		*lastMiddleClick = false
	}
}

func processMouseScroll(byteData *core.Bytedata) {
	if byteData.HasMouseScroll() {
		rotation := byteData.GetScrollRotation()
		fmt.Printf("Executing scroll with rotation: %d\n", rotation)
		if rotation > 0 {
			robotgo.Scroll(0, -3) // Scroll down
		} else if rotation < 0 {
			robotgo.Scroll(0, 3) // Scroll up
		}
	}
}
