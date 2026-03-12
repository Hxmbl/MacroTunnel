package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
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
	// Step 1: List local subnets
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	subnets := []string{}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ipParts := strings.Split(ipnet.IP.String(), ".")
				if len(ipParts) == 4 {
					subnet := fmt.Sprintf("%s.%s.%s", ipParts[0], ipParts[1], ipParts[2])
					subnets = append(subnets, subnet)
				}
			}
		}
	}

	// Remove duplicates
	seen := map[string]bool{}
	uniqueSubnets := []string{}
	for _, s := range subnets {
		if !seen[s] {
			seen[s] = true
			uniqueSubnets = append(uniqueSubnets, s)
		}
	}

	// Step 2: Ask user to pick subnet
	fmt.Println("Found local subnet candidates:")
	for i, s := range uniqueSubnets {
		fmt.Printf("%d) %s\n", i+1, s)
	}

	fmt.Print("Select the subnet your Mac is on [1-", len(uniqueSubnets), "]: ")
	var choice int
	fmt.Scan(&choice)
	if choice < 1 || choice > len(uniqueSubnets) {
		return nil, fmt.Errorf("invalid choice")
	}

	selectedSubnet := uniqueSubnets[choice-1]
	broadcastAddr := fmt.Sprintf("%s.255:9999", selectedSubnet)

	// Step 3: Send discovery ping
	remoteAddr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

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
