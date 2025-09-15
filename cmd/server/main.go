package main

// Minimal TCP server skeleton
// Accepts connections and sends a single welcome line, then closes the conn
// Will later turn this into a line-based protocol (login, deal, hit, stand, etc.)

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Bind address:
	// - Default is local only where we bind 127.0.0.1
	// - If user wants to run it on LAN we bind 0.0.0.0
	addr := "127.0.0.1:9090"
	if os.Getenv("LAN") == "1" {
		addr = "0.0.0.0:9090"
	}

	// Listen for TCP connections
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("Server listening on", ln.Addr().String(), "(stub)")

	// Main loop, for every incoming connection hand it to a goroutine
	for {
		conn, err := ln.Accept()
		if err != nil {
			// Accept can fail for a bit, log and continue
			log.Println("Accept error:", err)
			continue
		}

		// Handle each client concurrently so slow clients do not block others
		go func(c net.Conn) {
			defer c.Close()
			c.Write([]byte("Welcome to Casino (stub server)!\n"))
			// TODO:
			// 1 - Read newline terminated commands from the client (e.g. "login <username> <password>")
			// 2 - Route commands to security for authentication and core for game
			// 3 - Write responses back to the client (e.g., "OK ...", "ERROR ...")
			// 4 - Add timeouts to avoid hanging connections
		}(conn)
	}
}
