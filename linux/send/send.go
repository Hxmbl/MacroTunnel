package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Message struct {
	DeviceID string `json:"device_id"`
	KeyID    int    `json:"key_id"`
}

func main() {
	// Hardcoded macOS receiver IP for now
	serverIP := "192.168.1.42:9090"

	conn, err := net.Dial("tcp", serverIP)
	if err != nil {
		fmt.Println("Failed to connect:", err)
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
		// NDJSON: send JSON + newline
		writer.Write(jsonBytes)
		writer.WriteByte('\n')
		writer.Flush()

		fmt.Println("Sent:", string(jsonBytes))
		time.Sleep(500 * time.Millisecond) // simulate time between keypresses
	}
}
