package helpers

import (
	"encoding/json"
	"fmt"
	"net"
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
