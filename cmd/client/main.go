package main

// Minimal TCP client skeleton
// Reads one line, prints it, and exits
// Will later turn this into an interactive REPL that sends commands

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// Attempt to connect to local server, if it isn't running ask user if the server is active
	conn, err := net.Dial("tcp", "127.0.0.1:9090")

	if err != nil {
		fmt.Println("Server not running yet? Try `make run-server` in another terminal.")
		os.Exit(1)
	}
	defer conn.Close()

	// Read one line from server and show it
	r := bufio.NewReader(conn)
	line, _ := r.ReadString('\n')

	fmt.Print(line)
	// TODO:
	// 1 - Implement a simple REPL:
	//    - Read line from stdin
	//    - Send it to server
	//    - Read and print server's reply
	// 2 - Add core commands: SIGNUP/LOGIN, DEAL, HIT, STAND, BALANCE, STATS, QUIT
	// 3 - Make HOST/PORT configurable (env vars or flags) for LAN testing
	fmt.Println("Client stub connected.")
}
