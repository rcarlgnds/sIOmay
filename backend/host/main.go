package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
#cgo CXXFLAGS: -std=c++17 -I../vendor/asio/asio/include
#cgo LDFLAGS: -L. -lcontroller -lstdc++ -lws2_32 -luser32 -static
#include "../internal_lib/internal_lib.hpp"
*/
import "C"

func main() {
	fromIP := flag.String("from", "", "IP address of the controller (e.g., 10.22.65.133:8080)")
	flag.Parse()

	if *fromIP == "" {
		fmt.Println("Usage: -from <ip:port>")
		os.Exit(1)
	}

	// Parse the IP and port
	parts := strings.Split(*fromIP, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid format. Use: ip:port")
		os.Exit(1)
	}

	ip := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("Invalid port number: %s\n", parts[1])
		os.Exit(1)
	}
	fmt.Printf("Starting client for %s:%d\n", ip, port)

	C.startClientC(C.CString("239.255.0.1"), C.int(8080))
}