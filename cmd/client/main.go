package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// Get server address from environment in the future or use default
	serverAddr := os.Getenv("CASINO_SERVER")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9090"
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server at %s\n", serverAddr)
		fmt.Println("Is the server running? Try `make run-server` in another terminal.")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to Casino server at %s\n", serverAddr)
	fmt.Println("Type 'help' for available commands or 'quit' to exit.")
	fmt.Println()

	go readFromServer(conn)
	writeToServer(conn)
}

func readFromServer(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response := scanner.Text()
		fmt.Println(response)

		if strings.HasPrefix(response, "OK") || strings.HasPrefix(response, "ERROR") {
			fmt.Print("> ")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Connection to server lost:", err)
		os.Exit(1)
	}
}

func writeToServer(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			fmt.Print("> ")
			continue
		}

		if strings.ToUpper(input) == "QUIT" || strings.ToUpper(input) == "EXIT" {
			conn.Write([]byte("QUIT\n"))
			return
		}

		parts := strings.Fields(input)
		if len(parts) >= 1 {
			command := strings.ToUpper(parts[0])
			if (command == "LOGIN" || command == "SIGNUP") && len(parts) == 2 {
				username := parts[1]
				password, err := getPassword("Password: ")
				if err != nil {
					fmt.Println("ERROR: Failed to read password")
					fmt.Print("> ")
					continue
				}
				input = fmt.Sprintf("%s %s %s", command, username, password)
			}
		}

		// Send command to server
		_, err := conn.Write([]byte(input + "\n"))
		if err != nil {
			fmt.Println("Failed to send command:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Input error:", err)
	}
}

func getPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read password without echoing
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	if err != nil {
		return "", err
	}

	return string(password), nil
}
