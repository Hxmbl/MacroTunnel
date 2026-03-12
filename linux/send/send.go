package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// DiscoveryResponse is what the Mac responds with
type DiscoveryResponse struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func discoverMac() (*DiscoveryResponse, error) {
	// Step 1: List local IPv4 subnets
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

	fmt.Printf("Select the subnet your Mac is on [1-%d]: ", len(uniqueSubnets))
	var choice int
	fmt.Scan(&choice)
	if choice < 1 || choice > len(uniqueSubnets) {
		return nil, fmt.Errorf("invalid choice")
	}

	selectedSubnet := uniqueSubnets[choice-1]
	broadcastAddr := fmt.Sprintf("%s.255:9999", selectedSubnet)

	// Step 3: Prepare UDP connection
	remoteAddr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveUDPAddr("udp4", ":0") // random local port
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Step 4: Send discovery ping
	_, err = conn.WriteToUDP([]byte("MACRO_DISCOVERY"), remoteAddr)
	if err != nil {
		return nil, err
	}
	fmt.Println("Discovery ping sent from", conn.LocalAddr())

	// Step 5: Wait for Mac to respond
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
