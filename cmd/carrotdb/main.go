package main

import (
	"bufio"
	"carrotdb/internal/engine"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("🥕 Welcome to CarrotDB")
	fmt.Println("Type 'SET key value', 'GET key', 'DELETE key', or 'EXIT' to quit.")

	// Ensure the data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize the engine with a log file in the data directory
	db, err := engine.NewEngine("data/carrotdb.log")
	if err != nil {
		log.Fatalf("failed to start engine: %v", err)
	}
	defer db.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			fmt.Print("> ")
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "EXIT", "QUIT":
			fmt.Println("Goodbye! 🥕")
			return

		case "SET":
			if len(parts) < 3 {
				fmt.Println("Usage: SET <key> <value>")
			} else {
				key := parts[1]
				value := strings.Join(parts[2:], " ")
				if err := db.Put(key, value); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}

		case "GET":
			if len(parts) < 2 {
				fmt.Println("Usage: GET <key>")
			} else {
				key := parts[1]
				val, err := db.Get(key)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println(val)
				}
			}

		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: DELETE <key>")
			} else {
				key := parts[1]
				if err := db.Delete(key); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v\n", err)
	}
}
