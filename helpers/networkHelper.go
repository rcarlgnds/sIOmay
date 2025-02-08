package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/go-ping/ping"
	"net"
	"sIOmay/object"
	"sort"
	"strings"
	"sync"
	"time"
)

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

func pingIP(ip string, timeout time.Duration, wg *sync.WaitGroup, results chan<- object.Computer) {
	defer wg.Done()

	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return
	}

	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		return
	}

	stats := pinger.Statistics()
	if stats.PacketLoss == 0 {
		results <- object.Computer{IPAddress: ip, Status: "Available"}
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

	for i := 1; i <= 254; i++ {
		clientIP := fmt.Sprintf("%s.%d", networkPrefix, i)

		wg.Add(1)
		go pingIP(clientIP, 500*time.Millisecond, &wg, results)
		fmt.Println(clientIP)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for client := range results {
		clients = append(clients, client)
	}

	// Sort clients by IP
	sort.Slice(clients, func(i, j int) bool {
		return ipToInt(clients[i].IPAddress) < ipToInt(clients[j].IPAddress)
	})
	return clients
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
	}
}

func AcknowledgeClient(connection *net.UDPConn, clientAddress *net.UDPAddr) {
	message := "Registered for cursor updates"
	_, err := connection.WriteToUDP([]byte(message), clientAddress)
	if err != nil {
		fmt.Println("Error when sending acknowledgment:", err)
	}
}
