package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	helper "sIOmay/helpers"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"fyne.io/fyne/v2/widget"

	_ "embed"
)

/*
#cgo LDFLAGS: -L. -lcor -lstdc++ -lws2_32 -luser32 -static
#include "../backend/internal_lib/extern.hpp"
*/
import "C"

//go:embed assets/PsExec.exe
var psexecBytes []byte

//go:embed .env
var envBytes []byte


func SetAuthToken(token string) {
	authToken = token
}

func GetAuthToken() string {
	return authToken
}

func IsConnected() bool {
	connectionMutex.Lock()
	defer connectionMutex.Unlock()
	return isConnected
}

func GetConnectedClients() []string {
	connectionMutex.Lock()
	defer connectionMutex.Unlock()
	result := make([]string, len(connectedClients))
	copy(result, connectedClients)
	return result
}

const (
	ServerPort    = 8080
	SleepDuration = 1 * time.Millisecond  
	BufferSize    = 1024  
)

type CmdexTokenDTO struct {
	IPAddresses []string `json:"ipaddresses"`
	TaskID      string   `json:"task_id"`
	CommandLine string   `json:"commandLine"`
}
var (
	isConnected      = false
	currentClients   []string
	connectedClients []string  // Track actually connected clients
	serverConnection *net.UDPConn
	stopServerChan   chan bool
	connectionMutex  sync.Mutex
	serverWg        sync.WaitGroup
	authToken       string
)

func init(){
	fmt.Print("Starting server... (Init CF)\n")
	go func() {
		C.startSiomayServerC()
	}()
}
func InitConnectButton(selectedComputer *[]string) *widget.Button {
	button := widget.NewButton("Connect", nil)
	
	button.OnTapped = func() {
		connectionMutex.Lock()
		defer connectionMutex.Unlock()
		
		if !isConnected {
			// Initial connection
			fmt.Printf("Connecting to %v\n", *selectedComputer)
			if len(*selectedComputer) > 0 {
				currentClients = make([]string, len(*selectedComputer))
				copy(currentClients, *selectedComputer)
				connectedClients = make([]string, len(*selectedComputer))
				copy(connectedClients, *selectedComputer)
				serverWg.Add(1)
				RunServer(*selectedComputer)
				isConnected = true
				button.SetText("Disconnect All")
				button.Refresh() 
				fmt.Println("Button changed to Disconnect All")
			} else {
				fmt.Println("No computers selected.")
			}
		} else {
			// Check if we're adding new connections or disconnecting
			newConnections := []string{}
			for _, selected := range *selectedComputer {
				isAlreadyConnected := false
				for _, connected := range connectedClients {
					if selected == connected {
						isAlreadyConnected = true
						break
					}
				}
				if !isAlreadyConnected {
					newConnections = append(newConnections, selected)
				}
			}
			
			if len(newConnections) > 0 {
				// Adding new connections
				fmt.Printf("Adding new connections to %v\n", newConnections)
				AddNewConnections(newConnections)
			} else {
				// Disconnect all
				fmt.Println("Disconnect button clicked - starting disconnection...")
				button.SetText("Disconnecting...")
				button.Refresh() 
				
				go func() {
					DisconnectFromClients()
					connectionMutex.Lock()
					isConnected = false
					connectedClients = []string{}
					connectionMutex.Unlock()
					button.SetText("Connect")
					button.Refresh()
					fmt.Println("Button changed back to Connect")
				}()
			}
		}
	}
	
	return button
}

func AddNewConnections(newIPs []string) {
	fmt.Printf("Adding new connections to %v\n", newIPs)
	
	// Add new IPs to connected clients list
	connectedClients = append(connectedClients, newIPs...)
	currentClients = append(currentClients, newIPs...)
	
	// Run server for new connections
	startSiomayId := "015c382c-3b93-43e4-a501-6b7c7addc638"
	err := RunRuman(newIPs, startSiomayId)
	if err != nil {
		fmt.Printf("Error running Ruman for new connections: %v\n", err)
	}
	
	fmt.Printf("New connections added. Total connected clients: %v\n", connectedClients)
}

func DisconnectFromClients() {
	fmt.Println("Starting disconnection process...")
	fmt.Println("Server goroutine finished.")

	stopSiomayClient := "0078d9f9-e676-4fed-a89f-480a1a0ed45f"
	RunRuman(currentClients, stopSiomayClient)
	for _, remoteMachine := range currentClients {
		StopServerForIP(remoteMachine)
	}

	currentClients = nil
	connectedClients = nil
}


func RunServer(allowedIPs []string) {
	serverIP, err := helper.GetServerIP()
	if err != nil {
		fmt.Println("Error getting server IP:", err)
		return
	}
	
	if !IsPortAvailable(serverIP, 8080) {
		fmt.Println("Error: Port 8080 is still in use. Please wait a moment and try again.")
		return
	}

		startControl(allowedIPs)
}
func IsPortAvailable(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}
func startControl(allowedIPs []string) {
	defer serverWg.Done()
	startSiomayId := "015c382c-3b93-43e4-a501-6b7c7addc638"
	er := RunRuman(allowedIPs, startSiomayId)
	if er != nil {
		fmt.Printf("Error running Ruman: %v\n", er)
	}


	runtime.LockOSThread()
	
}
func RunRuman(ListIps []string, taskid string) error{
	token := GetAuthToken()
	url := "https://api-ruman.apps.slc.net/cmdex/exec-token"
	
	payload := CmdexTokenDTO{
		IPAddresses: ListIps,
		TaskID:     taskid,
		CommandLine: "",
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK  || resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Ruman API error: %s", body)
	}

	return nil
}
func StopServerForIP(ipStr string) {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		fmt.Println("Invalid IP format:", ipStr)
		return
	}

	var ipParts [4]C.int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println("Invalid IP part:", p)
			return
		}
		ipParts[i] = C.int(n)
	}
	C.sendStopCommandC((*C.int)(unsafe.Pointer(&ipParts[0])))
}