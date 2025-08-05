package controller

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	helper "sIOmay/helpers"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"

	_ "embed"
)

//go:embed assets/PsExec.exe
var psexecBytes []byte


func LoadCredentials() (username, password string, err error) {
	
	err = godotenv.Load()
	if err != nil {
		
		err = godotenv.Load("../.env")
		if err != nil {
			return "", "", fmt.Errorf("error loading .env file: %w", err)
		}
	}
	
	username = os.Getenv("PSEXEC_USERNAME")
	password = os.Getenv("PSEXEC_PASSWORD")
	
	if username == "" || password == "" {
		return "", "", fmt.Errorf("PSEXEC_USERNAME and PSEXEC_PASSWORD must be set in .env file")
	}
	
	return username, password, nil
}


func isConnectionClosedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		   strings.Contains(errStr, "connection reset") ||
		   strings.Contains(errStr, "broken pipe")
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
)
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
				stopServerChan = make(chan bool, 1)
				
				go RunServer(*selectedComputer)
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
	
	
	if serverConnection != nil {
		fmt.Println("Sending disconnect messages to clients...")
		disconnectMsg := []byte("DISCONNECT")
		for _, remoteMachine := range currentClients {
			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteMachine, ServerPort))
			if err == nil {
				_, writeErr := serverConnection.WriteToUDP(disconnectMsg, addr)
				if writeErr != nil {
					fmt.Printf("Error sending disconnect to %s: %v\n", remoteMachine, writeErr)
				} else {
					fmt.Printf("Sent disconnect message to %s\n", remoteMachine)
				}
			} else {
				fmt.Printf("Error resolving address for %s: %v\n", remoteMachine, err)
			}
		}
		
		fmt.Println("Waiting for clients to process disconnect...")
		time.Sleep(1 * time.Second)
	} else {
		fmt.Println("No server connection found")
	}
	
	
	if stopServerChan != nil {
		fmt.Println("Signaling server to stop...")
		select {
		case stopServerChan <- true:
			fmt.Println("Stop signal sent")
		default:
			fmt.Println("Stop channel full or closed")
		}
	} else {
		fmt.Println("No stop channel found")
	}
	
	
	if serverConnection != nil {
		fmt.Println("Closing server connection...")
		serverConnection.Close()
		serverConnection = nil
		fmt.Println("Server connection closed")
	}
	
	
	fmt.Println("Killing client processes on remote machines...")
	for _, remoteMachine := range currentClients {
		fmt.Printf("Attempting to kill client on %s...\n", remoteMachine)
		err := KillClientOnRemoteMachine(remoteMachine)
		if err != nil {
			fmt.Printf("Failed to kill client on %s: %v\n", remoteMachine, err)
		} else {
			fmt.Printf("Client killed on %s\n", remoteMachine)
		}
	}
	
	currentClients = nil
	fmt.Println("Disconnected from all clients")
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

	go func() {
		startControl(allowedIPs)
	}()
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
	
	
	connection, _, err := helper.StartServer(serverIP, ServerPort)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	
	
	connectionMutex.Lock()
	serverConnection = connection
	connectionMutex.Unlock()
	
	defer func() {
		connection.Close()
		connectionMutex.Lock()
		serverConnection = nil
		connectionMutex.Unlock()
	}()
	
	fmt.Printf("Server is listening on port %d (%s)\n", ServerPort, serverIP)
	
	clientAddresses := make(map[string]*net.UDPAddr)
	for _, ip := range allowedIPs {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, ServerPort))
		if err != nil {
			fmt.Printf("Failed to resolve UDP address for IP %s: %v\n", ip, err)
			continue
		}
		clientAddresses[ip] = addr
	}
	
	mouseData := helper.NewMouse()
	sendChan := make(chan bool, 100) 
	mouseData.ListenForMouseEventsWithCallback(func() {
		select {
		case sendChan <- true:
		default: 
		}
	})
	go func() {
		for {
			select {
			case <-sendChan:
				helper.SendMouseMessageToClients(mouseData, clientAddresses, connection)
			case <-stopServerChan:
				fmt.Println("Stopping mouse data sender...")
				return
			}
		}
	}()
	
	
	buffer := make([]byte, BufferSize)
	for {
		select {
		case <-stopServerChan:
			fmt.Println("Stopping server...")
			return
		default:
			
			connection.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			
			_, clientAddress, err := connection.ReadFromUDP(buffer)
			if err != nil {
				
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue 
				}
				
				if isConnectionClosedError(err) {
					fmt.Println("Server connection closed - stopping UDP read loop")
					return
				}
				fmt.Println("Error reading from UDP:", err)
				continue
			}
			
			helper.RegisterClient(clientAddress, clientAddresses)
			helper.AcknowledgeClient(connection, clientAddress)
		}
	}
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
