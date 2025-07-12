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

	// Memulai goroutine untuk mencetak status secara real-time
	go func() {
		for status := range statusChan {
			fmt.Println(status)
		}
	}()

	// Ping
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

	// Result
	for client := range results {
		clients = append(clients, client)
	}

	// Sorting berdasarkan IP
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
	// Convert mouse struck ke bentuk json
	messageBytes, err := json.Marshal(mouse)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send Message JSON ke semua host
	for _, clientAddress := range clientAddresses {
		_, err := connection.WriteToUDP(messageBytes, clientAddress)
		if err != nil {

			fmt.Println(err.Error())
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
	// Register client yang belum ke register
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
