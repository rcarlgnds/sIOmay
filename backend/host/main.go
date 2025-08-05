package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sIOmay/core"
	"strings"

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
	var lastX, lastY int
	var lastLeftClick, lastRightClick, lastMiddleClick bool

	C.MoveMouse()
	for {
		if !processMouseData(connection, &lastX, &lastY, &lastLeftClick, &lastRightClick, &lastMiddleClick) {
			break
		}
	}
	fmt.Println("Client disconnected and shutting down")
}

func processMouseData(connection *net.UDPConn, lastX, lastY *int, lastLeftClick, lastRightClick, lastMiddleClick *bool) bool {
	fmt.Println("Waiting for mouse data...")
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
	fmt.Printf("📥 Received %d bytes from server\n", n)
	data := buffer[:n]
	if n > 0 {
    fmt.Printf("Raw data received: %v\n", data)
    fmt.Printf("Hex representation: %02X\n", data)
}
	fmt.Printf("=== RECEIVED MOUSE DATA ===\n")		
	fmt.Printf("Raw bytes: %v\n", data)
	fmt.Printf("Hex: %X\n", data)
	// if n > 0 && string(data) == "DISCONNECT" {
	// 	fmt.Println("Received disconnect command from server")
	// 	return false
	// }

	if n == 7 {
		executeMouseActions(data, lastX, lastY, lastLeftClick, lastRightClick, lastMiddleClick)
	} else if n > 0 {
		fmt.Printf("Received unexpected data size: %d bytes (expected 7)\n", n)
	}

	return true
}

func executeMouseActions(data []byte, lastX, lastY *int, lastLeftClick, lastRightClick, lastMiddleClick *bool) {
	byteData := core.NewBytedata()
	for i := 0; i < 7; i++ {
		byteData.Bytes()[i] = data[i]
	}
	fmt.Printf("=== RECEIVED MOUSE DATA ===\n")
	fmt.Printf("Raw bytes: %v\n", data)
	fmt.Printf("Hex: %X\n", data)

	processMouseMovement(byteData, lastX, lastY)
	processMouseClicks(byteData, lastLeftClick, lastRightClick, lastMiddleClick)
	processMouseScroll(byteData)

	fmt.Println("=============================")
}

func processMouseMovement(byteData *core.Bytedata, lastX, lastY *int) {
	if byteData.HasMouseMove() {
		x, y := byteData.GetMousePosition()
		if int(x) != *lastX || int(y) != *lastY {
			fmt.Printf("Moving mouse to (%d, %d)\n", x, y)
			robotgo.Move(int(x), int(y))
			*lastX, *lastY = int(x), int(y)
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
