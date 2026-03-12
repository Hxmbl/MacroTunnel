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
	broadcastAddr := "192.168.1.255:9999"
	remoteAddr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return nil, err
	}

	// Bind to random local port (but SAME socket for send + receive)
	localAddr, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Enable broadcast
	err = conn.SetWriteBuffer(1024)
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteToUDP([]byte("MACRO_DISCOVERY"), remoteAddr)
	if err != nil {
		return nil, err
	}

	fmt.Println("Discovery ping sent from", conn.LocalAddr())

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
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
