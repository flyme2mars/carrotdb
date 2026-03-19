package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("🥕 CarrotDB CLI Client")
	fmt.Println("Connecting to localhost:6379...")

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatalf("failed to connect to server: %v (Is the server running?)", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Type commands (e.g., 'SET key value', 'GET key', 'EXIT').")

	scanner := bufio.NewScanner(os.Stdin)
	reader := bufio.NewReader(conn)

	fmt.Print("> ")
	for scanner.Scan() {
		input := scanner.Text()
		if strings.TrimSpace(input) == "" {
			fmt.Print("> ")
			continue
		}

		// Send command to server
		fmt.Fprintf(conn, "%s\n", input)

		// Read response from server
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("failed to read from server: %v", err)
		}

		fmt.Print(response)

		if strings.ToUpper(input) == "EXIT" || strings.ToUpper(input) == "QUIT" {
			return
		}

		fmt.Print("> ")
	}
}
