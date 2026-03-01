package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type DiscoveryResponse struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type Message struct {
	DeviceID string `json:"device_id"`
	KeyID    int    `json:"key_id"`
}

func main() {
	tcpInfo, err := discoverMac()
	if err != nil {
		fmt.Println("Discovery failed:", err) // What we get after 2 seconds of waiting for a response, fix this later.
		return
	}

	serverAddr := fmt.Sprintf("%s:%d", tcpInfo.IP, tcpInfo.Port)
	fmt.Println("Connecting to discovered Mac at", serverAddr)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect TCP:", err)
		return
	}
	defer conn.Close()
	writer := bufio.NewWriter(conn)

	// Example messages
	messages := []Message{
		{DeviceID: "linux-test-1", KeyID: 1},
		{DeviceID: "linux-test-1", KeyID: 2},
		{DeviceID: "linux-test-1", KeyID: 3},
	}

	for _, msg := range messages {
		jsonBytes, _ := json.Marshal(msg)
		writer.Write(jsonBytes)
		writer.WriteByte('\n')
		writer.Flush()
		fmt.Println("Sent:", string(jsonBytes))
		time.Sleep(500 * time.Millisecond)
	}
}

func discoverMac() (*DiscoveryResponse, error) {
	broadcastAddr := "255.255.255.255:9999"
	udpAddr, _ := net.ResolveUDPAddr("udp4", broadcastAddr)
	conn, _ := net.DialUDP("udp4", nil, udpAddr)
	defer conn.Close()

	// Send discovery ping
	conn.Write([]byte("MACRO_DISCOVERY"))
	fmt.Println("Discovery ping sent")

	// Listen for response
	listenAddr, _ := net.ResolveUDPAddr("udp4", ":0")
	listenConn, _ := net.ListenUDP("udp4", listenAddr)
	defer listenConn.Close()

	listenConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, _, err := listenConn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	var resp DiscoveryResponse
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		return nil, err
	}
	fmt.Println("Discovered Mac:", resp)
	return &resp, nil
}
