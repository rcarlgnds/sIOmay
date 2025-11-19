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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	_ "embed"
)

/*
#cgo LDFLAGS: -L. -lcgo_compatible -lstdc++ -lws2_32 -luser32 -static
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

func SetGlobalWindow(window fyne.Window) {
	globalWindow = window
}

func ShowWindow() {
	if globalWindow != nil {
		fmt.Println("Showing main window...")
		globalWindow.Show()
		globalWindow.RequestFocus()
	} else {
		fmt.Println("Global window reference not set")
	}
}

func HideWindow() {
	if globalWindow != nil {
		fmt.Println("Hiding main window...")
		globalWindow.Hide()
	}
}
func VerifyToken(token string) (string, error) {
	url := "https://api-ruman.apps.slc.net/cmdex/verif-token"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("token verification failed: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return string(body), nil
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
	connectedClients []string
	serverConnection *net.UDPConn
	stopServerChan   chan bool
	connectionMutex  sync.Mutex
	serverWg         sync.WaitGroup
	authToken        string
	cleanupDone      = false
	cleanupMutex     sync.Mutex
	globalWindow     fyne.Window
)

func init() {
	go func() {
		C.startSiomayServerC()
	}()
}

func InitConnectButton(selectedComputer *[]string) *widget.Button {
	return InitConnectButtonWithCallback(selectedComputer, nil)
}

func InitConnectButtonWithWindow(selectedComputer *[]string, window fyne.Window) *widget.Button {
	return InitConnectButtonWithCallbackAndWindow(selectedComputer, nil, window)
}

func InitConnectButtonWithCallback(selectedComputer *[]string, onConnectionChange func()) *widget.Button {
	return InitConnectButtonWithCallbackAndWindow(selectedComputer, onConnectionChange, nil)
}

func InitConnectButtonWithCallbackAndWindow(selectedComputer *[]string, onConnectionChange func(), window fyne.Window) *widget.Button {
	button := widget.NewButton("Connect", nil)

	button.OnTapped = func() {
		connectionMutex.Lock()
		defer connectionMutex.Unlock()

		if !isConnected {
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
				if window != nil {
					fmt.Println("Auto-minimizing window after successful connection...")
					window.Hide()
				}
			} else {
				fmt.Println("No computers selected.")
			}
		} else {
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

			fmt.Printf("Debug - Selected: %v, Connected: %v, New: %v\n", *selectedComputer, connectedClients, newConnections)

			if len(newConnections) > 0 {
				fmt.Printf("Adding new connections to %v\n", newConnections)
				AddNewConnections(newConnections)
				*selectedComputer = []string{}
				fmt.Println("Cleared selected computers after adding connections")
				if window != nil {
					fmt.Println("Auto-minimizing window after adding new connections...")
					window.Hide()
				}

				if onConnectionChange != nil {
					onConnectionChange()
				}
			} else {
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
					if window != nil {
						fmt.Println("Auto-showing window after disconnecting all clients...")
						window.Show()
					}
				}()
			}
		}
	}

	return button
}
func ForceDisconnectAllClients() {
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()

	if cleanupDone {
		return
	}

	fmt.Println("Force disconnecting all clients...")

	connectionMutex.Lock()
	clientsToDisconnect := make([]string, len(connectedClients))
	copy(clientsToDisconnect, connectedClients)
	connectionMutex.Unlock()

	if len(clientsToDisconnect) > 0 {
		fmt.Printf("Disconnecting from clients: %v\n", clientsToDisconnect)

		stopSiomayClient := "0078d9f9-e676-4fed-a89f-480a1a0ed45f"
		err := RunRuman(clientsToDisconnect, stopSiomayClient)
		if err != nil {
			fmt.Printf("Error sending stop command: %v\n", err)
		}

		for _, remoteMachine := range clientsToDisconnect {
			StopServerForIP(remoteMachine)
		}

		fmt.Println("All clients disconnected successfully")
	}

	connectionMutex.Lock()
	isConnected = false
	currentClients = nil
	connectedClients = nil
	connectionMutex.Unlock()

	cleanupDone = true
}

func AddNewConnections(newIPs []string) {
	fmt.Printf("Adding new connections to %v\n", newIPs)

	connectedClients = append(connectedClients, newIPs...)
	currentClients = append(currentClients, newIPs...)

	startSiomayId := "015c382c-3b93-43e4-a501-6b7c7addc638"
	err := RunRuman(newIPs, startSiomayId)
	if err != nil {
		fmt.Printf("Error running Ruman for new connections: %v\n", err)
		return
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
func RunRuman(ListIps []string, taskid string) error {
	token := GetAuthToken()
	url := "https://api-ruman.apps.slc.net/cmdex/exec-token"

	payload := CmdexTokenDTO{
		IPAddresses: ListIps,
		TaskID:      taskid,
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

	if resp.StatusCode != http.StatusOK || resp.StatusCode != http.StatusCreated {
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
