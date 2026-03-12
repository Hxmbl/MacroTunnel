import Foundation
import Network

@available(macOS 10.14, *)
func main() {
    // Start TCP listener for keystrokes
    startTCPListener()

    // Start UDP discovery responder
    startUDPDiscoveryResponder()

    dispatchMain() // keeps both running
}

// ---------------- TCP Listener ----------------
@available(macOS 10.14, *)
func startTCPListener() {
    let tcpPort: NWEndpoint.Port = 9090

    do {
        let listener = try NWListener(using: .tcp, on: tcpPort)
        print("TCP listener running on port \(tcpPort)")

        listener.newConnectionHandler = { connection in
            print("New TCP connection:", connection)
            connection.start(queue: .global())
            receive(on: connection)
        }

        listener.start(queue: .main)
    } catch {
        print("Failed to start TCP listener:", error)
    }
}

@available(macOS 10.14, *)
func receive(on connection: NWConnection) {
    connection.receive(minimumIncompleteLength: 1, maximumLength: 4096) { data, _, isComplete, error in
        if let data = data, !data.isEmpty,
           let text = String(data: data, encoding: .utf8) {
            let lines = text.split(separator: "\n")
            for line in lines {
                print("Raw line received:", line)
                if let jsonData = line.data(using: .utf8) {
                    do {
                        if let obj = try JSONSerialization.jsonObject(with: jsonData) as? [String: Any] {
                            print("Parsed JSON:", obj)
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
                }
            }
        }

        if let error = error {
            print("TCP connection error:", error)
        }

        if !isComplete {
            receive(on: connection)
        } else {
            print("TCP connection closed:", connection)
        }
    }
}

// ---------------- UDP Discovery ----------------
@available(macOS 10.14, *)
func startUDPDiscoveryResponder() {
    let port: NWEndpoint.Port = 9999
    let listener = try! NWListener(using: .udp, on: port)
    print("UDP discovery listener running on port \(port)")

    listener.newConnectionHandler = { connection in
        connection.start(queue: .global())
        handleUDPConnection(connection)   // 🔥 START the loop
    }

    listener.start(queue: .main)
}

@available(macOS 10.14, *)
func handleUDPConnection(_ connection: NWConnection) {
    connection.receiveMessage { data, _, _, error in
        if let data = data,
           let text = String(data: data, encoding: .utf8),
           text == "MACRO_DISCOVERY" {

            print("Discovery ping received")

            let localIP = getLocalIP() ?? "127.0.0.1"
            let response = ["ip": localIP, "port": 9090]

            if let jsonData = try? JSONSerialization.data(withJSONObject: response) {
                connection.send(content: jsonData, completion: .contentProcessed { _ in
                    print("Sent discovery response to sender")
                })
            }
        }

        if error == nil {
            handleUDPConnection(connection)   // 🔁 keep listening
        }
    }
}

// ---------------- Helper: Get local IP ----------------
func getLocalIP() -> String? {
    var address: String?
    var ifaddr : UnsafeMutablePointer<ifaddrs>?
    if getifaddrs(&ifaddr) == 0 {
        var ptr = ifaddr
        while ptr != nil {
            let flags = Int32(ptr!.pointee.ifa_flags)
            let addr = ptr!.pointee.ifa_addr.pointee
            if addr.sa_family == UInt8(AF_INET) && (flags & (IFF_UP|IFF_RUNNING|IFF_LOOPBACK)) == (IFF_UP|IFF_RUNNING) {
                var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                getnameinfo(ptr!.pointee.ifa_addr, socklen_t(addr.sa_len),
                            &hostname, socklen_t(hostname.count),
                            nil, socklen_t(0), NI_NUMERICHOST)
                address = String(cString: hostname)
                break
            }
            ptr = ptr!.pointee.ifa_next
        }
        freeifaddrs(ifaddr)
    }
    return address
}

main()