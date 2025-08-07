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
	
	if distance < 3 {
		robotgo.Move(toX, toY)
		return
	}
	
	// Reduce steps for faster movement during bursts
	steps := int(math.Max(4, math.Min(15, distance/8)))
	
	for i := 1; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		
		// Cubic easing for smooth movement
		progress = 1 - math.Pow(1-progress, 3)
		
		currentX := fromX + int(dx*progress)
		currentY := fromY + int(dy*progress)
		
		robotgo.Move(currentX, currentY)
		
		// No sleep for faster movement
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

	fmt.Println("Starting byte-based mouse control with multi-threaded processing...")
	
	// Create a smaller buffered channel to minimize latency while handling bursts
	mouseDataBuffer := make(chan []byte, 50) // Reduced buffer size for lower latency
	
	// Start the dedicated receiving thread
	go receiveMouseData(connection, mouseDataBuffer)
	
	// Process buffered data in the main thread with rate limiting
	processBufferedMouseData(mouseDataBuffer, &lastX, &lastY, &lastLeftClick, &lastRightClick, &lastMiddleClick, &lastMoveTime)
	
	fmt.Println("Client disconnected and shutting down")
}

// Dedicated receiving thread - only focuses on receiving UDP packets
func receiveMouseData(connection *net.UDPConn, mouseDataBuffer chan<- []byte) {
	buffer := make([]byte, 1024)
	
	defer close(mouseDataBuffer) // Ensure channel is closed when function exits
	
	for {
		// No timeout - this thread is dedicated to receiving
		n, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			if isConnectionClosed(err) {
				fmt.Println("Server connection closed - stopping receiver thread")
				return
			}
			fmt.Println("Error reading from server:", err)
			continue
		}

		data := buffer[:n]

		if n > 0 && string(data) == "DISCONNECT" {
			fmt.Println("Received disconnect command from server")
			return
		}

		// Only process valid 7-byte mouse data
		if n == 7 {
			// Make a copy of the data to avoid race conditions
			dataCopy := make([]byte, 7)
			copy(dataCopy, data)
			
			// Try to send to buffer, drop packet if buffer is full
			select {
			case mouseDataBuffer <- dataCopy:
				// Successfully buffered
			default:
				// Buffer full - drop this packet to prevent blocking the receiver
				// This ensures we always process the latest mouse data
			}
		} else if n > 0 {
			fmt.Printf("Received unexpected data size: %d bytes (expected 7)\n", n)
		}
	}
}

// Process buffered mouse data from the receiving thread
func processBufferedMouseData(mouseDataBuffer <-chan []byte, lastX, lastY *int, lastLeftClick, lastRightClick, lastMiddleClick *bool, lastMoveTime *time.Time) {
	for data := range mouseDataBuffer {
		if data == nil {
			break // Channel closed
		}
		
		executeMouseActions(data, lastX, lastY, lastLeftClick, lastRightClick, lastMiddleClick, lastMoveTime)
		
		// Small sleep to prevent overwhelming the system while still being responsive
		time.Sleep(time.Microsecond * 500) // ~2000 FPS max processing rate
	}
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
		
		// Reduced rate limiting: allow up to 250 times per second
		now := time.Now()
		if now.Sub(*lastMoveTime) < time.Millisecond*4 {
			return
		}
		*lastMoveTime = now
		
		// Only move if position actually changed
		if newX != *lastX || newY != *lastY {
			// fmt.Printf("Moving mouse from (%d, %d) to (%d, %d)\n", *lastX, *lastY, newX, newY)
			
			// Use direct movement for faster processing during bursts
			distance := math.Sqrt(float64((newX-*lastX)*(newX-*lastX) + (newY-*lastY)*(newY-*lastY)))
			if distance > 50 {
				// Only use smooth movement for large jumps
				smoothMove(*lastX, *lastY, newX, newY)
			} else {
				// Direct movement for small distances to reduce processing time
				robotgo.Move(newX, newY)
			}
			
			*lastX, *lastY = newX, newY
		}
	}
}

func processMouseClicks(byteData *core.Bytedata, lastLeftClick, lastRightClick, lastMiddleClick *bool) {
	// Left click processing
	if byteData.HasMouseClickLeft() && !*lastLeftClick {
		fmt.Println(">>> CLIENT: Executing left click")
		robotgo.MouseClick("left", false)
		*lastLeftClick = true
	} else if !byteData.HasMouseClickLeft() && *lastLeftClick {
		fmt.Println(">>> CLIENT: Left click released")
		*lastLeftClick = false
	}

	// Right click processing
	if byteData.HasMouseClickRight() && !*lastRightClick {
		fmt.Println(">>> CLIENT: Executing right click")
		robotgo.MouseClick("right", false)
		*lastRightClick = true
	} else if !byteData.HasMouseClickRight() && *lastRightClick {
		fmt.Println(">>> CLIENT: Right click released")
		*lastRightClick = false
	}

	// Middle click processing
	if byteData.HasMouseMiddleClick() && !*lastMiddleClick {
		fmt.Println(">>> CLIENT: Executing middle click")
		robotgo.MouseClick("center", false)
		*lastMiddleClick = true
	} else if !byteData.HasMouseMiddleClick() && *lastMiddleClick {
		fmt.Println(">>> CLIENT: Middle click released")
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
