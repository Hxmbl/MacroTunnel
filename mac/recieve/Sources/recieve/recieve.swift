// The Swift Programming Language
// https://docs.swift.org/swift-book

import Foundation
import Network

@available(macOS 10.14, *)
func main() { // Start the server
    let port: NWEndpoint.Port = 9090

    do {
        let listener = try NWListener(using: .tcp, on: port)
        print("Listening on port 9090")

        listener.newConnectionHandler = { connection in
            print("New connection established:", connection)
            connection.start(queue: .global())
            receive(on: connection)
        }

        listener.start(queue: .main)
        dispatchMain()
    } catch {
        print("Failed to start listener:", error)
    }
}

@available(macOS 10.14, *)
func receive(on connection: NWConnection) {
    connection.receive(minimumIncompleteLength: 1, maximumLength: 4096) { data, _, isComplete, error in
        if let data = data, !data.isEmpty,
           let text = String(data: data, encoding: .utf8) {

            // Split by newline for NDJSON
            let lines = text.split(separator: "\n")
            for line in lines {
                print("Raw line received:", line)

                // Attempt to parse JSON
                if let jsonData = line.data(using: .utf8) {
                    do {
                        if let obj = try JSONSerialization.jsonObject(with: jsonData) as? [String: Any] {
                            // Verbose: print parsed object
                            print("Parsed JSON:", obj)

                            // Optional: extract expected keys
                            if let device = obj["device_id"] as? String,
                               let key = obj["key_id"] as? Int {
                                print("Key pressed:", key, "from device:", device)
                            } else {
                                print("JSON missing expected keys, ignoring:", obj)
                            }

                        } else {
                            print("JSON is not a dictionary, ignoring:", line)
                        }
                    } catch {
                        print("Discarded invalid JSON:", line)
                    }
                } else {
                    print("Failed to convert line to Data, ignoring:", line)
                }
            }
        }

        if let error = error {
            print("Connection error:", error)
        }

        if !isComplete {
            // Recursively keep receiving
            receive(on: connection)
        } else {
            print("Connection closed:", connection)
        }
    }
}

main()