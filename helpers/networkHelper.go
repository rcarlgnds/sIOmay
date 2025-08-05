package helpers

import (
	"encoding/json"
	"fmt"
	"net"

	"sIOmay/object"
	"sort"
	"strings"
	"sync"
	"time"

	// controller "sIOmay/controller"
	"github.com/go-ping/ping"
)

// Statistics for byte transmission tracking
var (
	totalBytesSent int64
	totalPacketsSent int64
	statisticsMutex sync.Mutex
	verboseByteLogging bool = true // Set to false to reduce console spam
)

// func RunServer() {
// 	cmd := exec.Command("go", "run", "backend/server/main.go")

// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	// Pake Start() biar gk ngehalangin kode lain
// 	// Waktu pake Run() langsung ke disable guinya wokwokwowk
// 	err := cmd.Start()
// 	if err != nil {
// 		fmt.Println("Error starting server: ", err)
// 		return
// 	}

// 	fmt.Println("Server started")
// }

func GetServerIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addresses {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}

func GetNetworkPrefix(ip string) string {
	octets := strings.Split(ip, ".")
	if len(octets) < 3 {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s", octets[0], octets[1], octets[2])
}

func pingIP(ip string, timeout time.Duration, wg *sync.WaitGroup, results chan<- object.Computer, statusChan chan<- string) {
	defer wg.Done()

	pinger, err := ping.NewPinger(ip)
	if err != nil {
		statusChan <- fmt.Sprintf("[✘] %s: Error initializing pinger", ip)
		return
	}

	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		statusChan <- fmt.Sprintf("[✘] %s: Ping failed", ip)
		return
	}

	stats := pinger.Statistics()
	if stats.PacketLoss == 0 {
		results <- object.Computer{IPAddress: ip, Status: "Available"}
		statusChan <- fmt.Sprintf("[✔] %s: Available", ip)
	} else {
		statusChan <- fmt.Sprintf("[✘] %s: Unreachable", ip)
	}
}

func ipToInt(ip string) int {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0
	}
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0
	}
	return int(ipv4[0])<<24 | int(ipv4[1])<<16 | int(ipv4[2])<<8 | int(ipv4[3])
}

func GetAllClients(serverIP string) []object.Computer {
	var clients []object.Computer
	networkPrefix := GetNetworkPrefix(serverIP)
	if networkPrefix == "" {
		fmt.Println("Invalid Server IP")
		return clients
	}

	var wg sync.WaitGroup
	results := make(chan object.Computer, 255)
	statusChan := make(chan string, 255) // Channel untuk cetak status ping

	go func() {
		for status := range statusChan {
			fmt.Println(status)
		}
	}()

	for i := 1; i <= 254; i++ {
		clientIP := fmt.Sprintf("%s.%d", networkPrefix, i)

		wg.Add(1)
		go pingIP(clientIP, 500*time.Millisecond, &wg, results, statusChan)
	}

	go func() {
		wg.Wait()
		close(results)
		close(statusChan)
	}()

	for client := range results {
		clients = append(clients, client)
	}

	sort.Slice(clients, func(i, j int) bool {
		return ipToInt(clients[i].IPAddress) < ipToInt(clients[j].IPAddress)
	})

	return clients
}


func SendKeyboardMessageToClients(keyboard *Keyboard, clientAddresses map[string]*net.UDPAddr, connection *net.UDPConn) {
	messageBytes, err := json.Marshal(keyboard)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, clientAddress := range clientAddresses {
		_, err := connection.WriteToUDP(messageBytes, clientAddress)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func SendMouseMessageToClients(mouse *Mouse, clientAddresses map[string]*net.UDPAddr, connection *net.UDPConn) {
	// Check if there's new byte data to send
	if !mouse.HasNewByteData() {
		return // No new data to send
	}
	
	// Get the byte data (7 bytes)
	byteData := mouse.GetByteData()
	if byteData == nil {
		return
	}

	// Print byte information if verbose logging is enabled
	if verboseByteLogging {
		printDetailedByteInfo(byteData)
	}

	// Send to all clients and collect statistics
	clientCount := sendBytesToClients(byteData, clientAddresses, connection)
	updateAndPrintStatistics(byteData, clientCount, len(clientAddresses))
	
	// Clear the byte data after sending
	mouse.ClearByteData()
}

func printDetailedByteInfo(byteData []byte) {
	fmt.Printf("=== SENDING MOUSE DATA ===\n")
	fmt.Printf("Raw bytes: %v\n", byteData)
	fmt.Printf("Hex representation: %02X\n", byteData)
	fmt.Printf("Binary representation: ")
	for i, b := range byteData {
		fmt.Printf("[%d]=%08b ", i, b)
	}
	fmt.Println()
	
	if len(byteData) >= 7 {
		printDecodedByteData(byteData)
	}
}

func printDecodedByteData(byteData []byte) {
	opcode := byteData[0]
	specialKey := byteData[1]
	x := int16(byteData[2])<<8 | int16(byteData[3])
	y := int16(byteData[4])<<8 | int16(byteData[5])
	
	fmt.Printf("Decoded data:\n")
	fmt.Printf("  Opcode: 0x%02X (%08b)\n", opcode, opcode)
	fmt.Printf("  Special Key: 0x%02X\n", specialKey)
	fmt.Printf("  X coordinate: %d\n", x)
	fmt.Printf("  Y coordinate: %d\n", y)
	
	printOperationDetails(opcode, x, y)
}

func printOperationDetails(opcode byte, x, y int16) {
	fmt.Printf("  Operations detected:\n")
	if opcode&0x08 != 0 { // MouseClickLeft = 0b00001000
		fmt.Printf("    - Left Click\n")
	}
	if opcode&0x04 != 0 { // MouseClickRight = 0b00000100
		fmt.Printf("    - Right Click\n")
	}
	if opcode&0x01 != 0 { // MouseMiddleClick = 0b00000001
		fmt.Printf("    - Middle Click\n")
	}
	if opcode&0x10 != 0 { // MouseMove = 0b00010000
		fmt.Printf("    - Mouse Move to (%d, %d)\n", x, y)
	}
	if opcode&0x02 != 0 { // MouseScroll = 0b00000010
		fmt.Printf("    - Mouse Scroll (rotation: %d)\n", x) // x contains rotation for scroll
	}
}

func sendBytesToClients(byteData []byte, clientAddresses map[string]*net.UDPAddr, connection *net.UDPConn) int {
	clientCount := 0
	for clientIP, clientAddress := range clientAddresses {
		_, err := connection.WriteToUDP(byteData, clientAddress)
		if err != nil {
			fmt.Printf("❌ Error sending to client %s (%s): %v\n", clientIP, clientAddress, err)
		} else {
			clientCount++
			if verboseByteLogging {
				fmt.Printf("✅ Sent %d bytes to client %s (%s)\n", len(byteData), clientIP, clientAddress)
			}
		}
	}
	return clientCount
}

func updateAndPrintStatistics(byteData []byte, clientCount, totalClients int) {
	bytesThisPacket := int64(len(byteData))
	
	// Update statistics
	statisticsMutex.Lock()
	totalBytesSent += bytesThisPacket * int64(clientCount)
	totalPacketsSent++
	currentTotalBytes := totalBytesSent
	currentTotalPackets := totalPacketsSent
	statisticsMutex.Unlock()
	
	if verboseByteLogging {
		fmt.Printf("Total clients sent to: %d/%d\n", clientCount, totalClients)
		fmt.Printf("📊 Session stats: %d packets sent, %d total bytes sent\n", currentTotalPackets, currentTotalBytes)
		fmt.Printf("==========================\n\n")
	} else {
		// Show compact summary every 100 packets
		if currentTotalPackets%100 == 0 {
			fmt.Printf("📊 Sent packet #%d (%d bytes) to %d clients\n", currentTotalPackets, bytesThisPacket, clientCount)
		}
	}
}
func StartServer(serverIP string, serverPort int) (*net.UDPConn, *net.UDPAddr, error) {
	address := net.UDPAddr{
		IP:   net.ParseIP(serverIP),
		Port: serverPort,
	}

	connection, err := net.ListenUDP("udp4", &address)
	if err != nil {
		fmt.Println("Error when starting server: ", err)
	}
	return connection, &address, err
}

func RegisterClient(clientAddress *net.UDPAddr, clientAddresses map[string]*net.UDPAddr) {
	clientKey := clientAddress.String()
	if _, exists := clientAddresses[clientKey]; !exists {
		fmt.Printf("New client registered: %s\n", clientKey)
		clientAddresses[clientKey] = clientAddress
	} else {
		fmt.Printf("Client %s already registered\n", clientKey)
	}
	fmt.Printf("Total registered clients: %d\n", len(clientAddresses))
}

func AcknowledgeClient(connection *net.UDPConn, clientAddress *net.UDPAddr) {
	message := "Registered for cursor updates"
	_, err := connection.WriteToUDP([]byte(message), clientAddress)
	if err != nil {
		fmt.Println("Error when sending acknowledgment:", err)
	}
}

func PrintByteTransmissionStats() {
	statisticsMutex.Lock()
	defer statisticsMutex.Unlock()
	
	fmt.Printf("\n📈 === BYTE TRANSMISSION STATS ===\n")
	fmt.Printf("Total packets sent: %d\n", totalPacketsSent)
	fmt.Printf("Total bytes sent: %d\n", totalBytesSent)
	if totalPacketsSent > 0 {
		fmt.Printf("Average bytes per packet: %.2f\n", float64(totalBytesSent)/float64(totalPacketsSent))
	}
	fmt.Printf("===================================\n\n")
}

// ResetByteTransmissionStats resets the transmission statistics
func ResetByteTransmissionStats() {
	statisticsMutex.Lock()
	defer statisticsMutex.Unlock()
	
	totalBytesSent = 0
	totalPacketsSent = 0
	fmt.Println("📊 Byte transmission statistics reset")
}

// SetVerboseByteLogging enables or disables verbose byte logging
func SetVerboseByteLogging(enabled bool) {
	verboseByteLogging = enabled
	if enabled {
		fmt.Println("🔍 Verbose byte logging ENABLED - detailed byte information will be shown")
	} else {
		fmt.Println("🔇 Verbose byte logging DISABLED - showing summary every 100 packets")
	}
}

// IsVerboseByteLogging returns the current verbose logging state
func IsVerboseByteLogging() bool {
	return verboseByteLogging
}
