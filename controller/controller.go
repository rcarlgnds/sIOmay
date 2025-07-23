package controller

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	helper "sIOmay/helpers"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"

	_ "embed"
)

//go:embed assets/PsExec.exe
var psexecBytes []byte

const (
	ServerPort    = 8080
	SleepDuration = 10 * time.Millisecond
	BufferSize    = 256
)
func InitConnectButton(selectedComputer *[]string) *widget.Button {
	return widget.NewButton("Connect", func() {
		fmt.Printf("Connecting to %v\n", *selectedComputer)
		if len(*selectedComputer) > 0 {
			go RunServer(*selectedComputer)
		} else {
			fmt.Println("No computers selected.")
		}
	})
}
func RunServer(allowedIPs []string) {

	serverIP, err := helper.GetServerIP()
	if err != nil {
		fmt.Println("Error getting server IP:", err)
	}
	
	if !IsPortAvailable(serverIP, 8080) {
		fmt.Println("Error: Port 8080 is still in use. Please wait a moment and try again.")
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
	for _, remoteMachine := range allowedIPs { // jangan dibuka 
		username := "" // username admin
		password := "" //password admin
	
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
	defer connection.Close()
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
	var mouseMu sync.Mutex
	mouseData := helper.NewMouse()
	mouseData.ListenForMouseEvents()
	go func() {
		go func() {
			for {
				mouseMu.Lock()
				if mouseData.HasMouseChanged() {
				}
				mouseMu.Unlock()
				time.Sleep(SleepDuration)
			}
		}()
		for {
			mouseData.Mu.Lock()
			fmt.Printf("Mouse Data: %+v\n", mouseData)
			helper.SendMouseMessageToClients(mouseData, clientAddresses, connection)
			mouseData.Mu.Unlock()
			time.Sleep(SleepDuration)
		}
	}()
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
	fmt.Println("Running PsExec:", cmd.String())
	return cmd.Start()
}
