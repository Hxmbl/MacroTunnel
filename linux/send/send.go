package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
		fmt.Println("Discovery failed:", err)
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

	fmt.Println("Type keys to send. Only single characters will be sent; strings ignored.")

	// Read from terminal
	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		// Ignore newline
		if r == '\n' || r == '\r' {
			continue
		}

		keyID := mapRuneToKeyID(r)
		if keyID == 0 {
			fmt.Printf("Ignored character: %c\n", r)
			continue
		}

		msg := Message{
			DeviceID: "linux-test-1",
			KeyID:    keyID,
		}

		jsonBytes, _ := json.Marshal(msg)
		writer.Write(jsonBytes)
		writer.WriteByte('\n')
		writer.Flush()
		fmt.Println("Sent:", string(jsonBytes))
		time.Sleep(50 * time.Millisecond) // optional small delay
	}
}

// Map terminal rune to your key IDs
func mapRuneToKeyID(r rune) int {
	switch r {
	case '1':
		return 1
	case '2':
		return 2
	case '3':
		return 3
	case '4':
		return 4
	default:
		return 0 // ignore everything else
	}
}

// --- Discovery logic remains unchanged ---
func discoverMac() (*DiscoveryResponse, error) {
	// ... your existing discoverMac() code here ...
	return nil, nil // placeholder
}
