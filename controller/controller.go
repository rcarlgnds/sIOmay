package controller

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	helper "sIOmay/helpers"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"

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

func LoadCredentials() (username, password string, err error) {
	envMap, err := godotenv.Unmarshal(string(envBytes))
	if err != nil {
		return "", "", fmt.Errorf("error parsing embedded .env content: %w", err)
	}

	for key, value := range envMap {
		os.Setenv(key, value)
	}

	username = os.Getenv("PSEXEC_USERNAME")
	password = os.Getenv("PSEXEC_PASSWORD")

	if username == "" || password == "" {
		return "", "", fmt.Errorf("PSEXEC_USERNAME and PSEXEC_PASSWORD must be set in .env file")
	}

	return username, password, nil
}



const (
	ServerPort    = 8080
	SleepDuration = 1 * time.Millisecond  
	BufferSize    = 1024  
)


var (
	isConnected      = false
	currentClients   []string
	serverConnection *net.UDPConn
	stopServerChan   chan bool
	connectionMutex  sync.Mutex
	serverWg        sync.WaitGroup
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
			
			fmt.Printf("Connecting to %v\n", *selectedComputer)
			if len(*selectedComputer) > 0 {
				currentClients = make([]string, len(*selectedComputer))
				copy(currentClients, *selectedComputer)
				serverWg.Add(1)
				RunServer(*selectedComputer)
				isConnected = true
				button.SetText("Disconnect")
				button.Refresh() 
				fmt.Println("Button changed to Disconnect")
			} else {
				fmt.Println("No computers selected.")
			}
		} else {
			
			fmt.Println("Disconnect button clicked - starting disconnection...")
			button.SetText("Disconnecting...")
			button.Refresh() 
			
			go func() {
				DisconnectFromClients()

				
					connectionMutex.Lock()
					isConnected = false
					connectionMutex.Unlock()
					button.SetText("Connect")
					button.Refresh()
					fmt.Println("Button changed back to Connect")
			}()
		}
	}
	
	return button
}

func DisconnectFromClients() {
	fmt.Println("Starting disconnection process...")
	fmt.Println("Server goroutine finished.")

		fmt.Println("Killing client processes on remote machines...")
	for _, remoteMachine := range currentClients {
		StopServerForIP(remoteMachine)
		fmt.Printf("Attempting to kill client on %s...\n", remoteMachine)
		err := KillClientOnRemoteMachine(remoteMachine)
		if err != nil {
			fmt.Printf("Failed to kill client on %s: %v\n", remoteMachine, err)
		} else {
			fmt.Printf("Client killed on %s\n", remoteMachine)
		}
	}

	currentClients = nil
}

func KillClientOnRemoteMachine(remoteMachine string) error {
	tmpDir := os.TempDir()
	psExecPath := filepath.Join(tmpDir, "PsExec.exe")
	
	
	username, password, err := LoadCredentials()
	if err != nil {
		return fmt.Errorf("error loading credentials: %w", err)
	}
	
	fmt.Printf("Attempting to kill client.exe on %s...\n", remoteMachine)
	
	
	cmd := exec.Command(
		psExecPath,
		"-accepteula",
		"-u", username,
		"-p", password,
		"\\\\"+remoteMachine,
		"taskkill", "/f", "/im", "client.exe",
	)
	
	
	output, err := cmd.CombinedOutput()
	fmt.Printf("PsExec kill output for %s: %s\n", remoteMachine, string(output))
	
	if err != nil {
		fmt.Printf("PsExec kill error for %s: %v\n", remoteMachine, err)
		
		
		fmt.Printf("Trying alternative kill method for %s...\n", remoteMachine)
		altCmd := exec.Command("taskkill", "/s", remoteMachine, "/u", username, "/p", password, "/f", "/im", "client.exe")
		altOutput, altErr := altCmd.CombinedOutput()
		fmt.Printf("Alternative kill output for %s: %s\n", remoteMachine, string(altOutput))
		if altErr != nil {
			fmt.Printf("Alternative kill error for %s: %v\n", remoteMachine, altErr)
		}
		return altErr
	}
	
	fmt.Printf("Successfully killed client on %s\n", remoteMachine)
	return nil
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
	serverIP, err := helper.GetServerIP()
	if err != nil {
		panic(err)
	}
	username, password, err := LoadCredentials()
	if err != nil {
		fmt.Printf("Error loading credentials: %v\n", err)
		fmt.Println("Please create a .env file with PSEXEC_USERNAME and PSEXEC_PASSWORD")
		return
	}

	for _, remoteMachine := range allowedIPs {
		err := RunClientWithPsExec(serverIP, remoteMachine, username, password)
		if err != nil {
			fmt.Printf("Failed to run PsExec on %s: %v\n", remoteMachine, err)
		} else {
			fmt.Printf("PsExec started on %s with IP %s\n", remoteMachine, serverIP)
		}
	}

	runtime.LockOSThread()
	
}
func RunClientWithRuman(){
	
}
func RunClientWithPsExec(serverIP, remoteMachine, username, password string) error {
	tmpDir := os.TempDir()
	psExecPath := filepath.Join(tmpDir, "PsExec.exe")
	if _, err := os.Stat(psExecPath); os.IsNotExist(err) {
		err := os.WriteFile(psExecPath, psexecBytes, 0755)
		if err != nil {
			return fmt.Errorf("failed to write PsExec: %w", err)
		}
	}
	clientPath := "C:\\Program Files\\client\\client.exe"
	cmd := exec.Command(
		psExecPath,
		"-accepteula",
		"-i", "1",
		"-u", username,
		"-p", password,
		"\\\\"+remoteMachine,
		clientPath,
		"-from", fmt.Sprintf("%s:%d", serverIP, ServerPort),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
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