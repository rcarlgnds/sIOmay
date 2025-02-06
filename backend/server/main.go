package main

import (
	"fmt"
	"net"
	helper "sIOmay/helpers"
	"sync"
	"time"
)

const (
	ServerPort    = 8080
	SleepDuration = 10 * time.Millisecond
	BufferSize    = 1024
)

func main() {
	// Get Server IP
	serverIP, err := helper.GetServerIP()
	if err != nil {
		panic(err)
	}

	// Start Server
	connection, _, err := helper.StartServer(serverIP, ServerPort)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer func(connection *net.UDPConn) {
		err := connection.Close()
		if err != nil {

		}
	}(connection)
	fmt.Printf("Server is listening on port %d (%s)\n", ServerPort, serverIP)

	// Mapping IP Host
	clientAddresses := make(map[string]*net.UDPAddr)

	// Create synchronization objects
	//var keyboardMu sync.Mutex
	var mouseMu sync.Mutex
	//var mouseDataChanged bool
	//var keyboardDataChanged bool

	// Create mouse event object
	mouseData := helper.NewMouse()
	mouseData.ListenForMouseEvents()

	// Keyboard event handling
	//events := make(chan *helper.Keyboard)
	//go func() {
	//	keyboardData := helper.NewKeyboardEvent()
	//	err := keyboardData.ListenForGlobalKeyboardEvents(events)
	//	if err != nil {
	//		fmt.Println("Error listening for keyboard events:", err)
	//	}
	//}()

	// Main processing goroutine
	go func() {
		//var keyboardData *helper.Keyboard

		// Goroutine to update keyboardData from events
		//go func() {
		//	for event := range events {
		//		keyboardMu.Lock()
		//		keyboardData = event
		//		keyboardDataChanged = true
		//		keyboardMu.Unlock()
		//	}
		//}()

		// Goroutine to update mouseData when mouse events are detected
		go func() {
			for {
				mouseMu.Lock()
				if mouseData.HasMouseChanged() {
					//mouseDataChanged = true
				}
				mouseMu.Unlock()
				time.Sleep(SleepDuration)
			}
		}()

		// Mouse Send
		for {
			mouseData.Mu.Lock()
			fmt.Printf("Mouse Data: %+v\n", mouseData)
			//if mouseDataChanged {
			//	fmt.Printf("Mouse Data: %+v\n", mouseData)
			//	mouseDataChanged = false
			helper.SendMouseMessageToClients(mouseData, clientAddresses, connection)
			//}
			mouseData.Mu.Unlock()
			time.Sleep(SleepDuration)
		}

		//for {
		//	keyboardMu.Lock()
		//	if keyboardDataChanged {
		//		fmt.Printf("Keyboard Data: %+v\n", keyboardData)
		//		mouseDataChanged = false
		//		helper.SendKeyboardMessageToClients(keyboardData, clientAddresses, connection)
		//	}
		//	keyboardData.Mu.Unlock()
		//	time.Sleep(SleepDuration)
		//}

	}()

	// Handle UDP communication
	buffer := make([]byte, BufferSize)
	for {
		_, clientAddress, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		helper.RegisterClient(clientAddress, clientAddresses)
		helper.AcknowledgeClient(connection, clientAddress)
	}
}
