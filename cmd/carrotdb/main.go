package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	host     = flag.String("host", "localhost", "CarrotDB Router host")
	port     = flag.String("port", "8000", "CarrotDB Router port")
	helpFlag = flag.Bool("help", false, "Show help")
)

func main() {
	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	addr := fmt.Sprintf("%s:%s", *host, *port)

	// Mode Selection
	args := flag.Args()
	if len(args) > 0 {
		// One-Shot Mode
		executeOneShot(addr, strings.Join(args, " "))
	} else {
		// Interactive REPL Mode
		startREPL(addr)
	}
}

func executeOneShot(addr, command string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		color.Red("Error: Could not connect to CarrotDB at %s", addr)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\n", command)
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		color.Red("Error: Failed to read from server: %v", err)
		os.Exit(1)
	}

	formatResponse(command, response)
}

func startREPL(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		color.Red("Error: Could not connect to CarrotDB at %s", addr)
		color.Yellow("Hint: Is the CarrotDB server/router running?")
		os.Exit(1)
	}
	defer conn.Close()

	color.Cyan("🥕 CarrotDB CLI Client (v0.14.0)")
	color.Green("Connected to %s", addr)
	fmt.Println("Type 'HELP' for commands or 'EXIT' to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	reader := bufio.NewReader(conn)

	prompt := color.New(color.FgHiBlue).SprintFunc()
	
	for {
		fmt.Printf("%s> ", prompt("carrot["+addr+"]"))
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		cmdUpper := strings.ToUpper(input)
		if cmdUpper == "EXIT" || cmdUpper == "QUIT" {
			color.Yellow("Goodbye!")
			return
		}

		if cmdUpper == "HELP" {
			printInteractiveHelp()
			continue
		}

		// Send command to server
		fmt.Fprintf(conn, "%s\n", input)

		// Read response
		response, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Disconnected: %v", err)
			return
		}

		formatResponse(input, response)
	}
}

func formatResponse(command, response string) {
	response = strings.TrimSpace(response)
	if response == "" {
		return
	}

	// Handle Success/Error prefixes
	if strings.HasPrefix(response, "+OK") {
		color.Green(response)
	} else if strings.HasPrefix(response, "+") {
		// Data response (GET, KEYS, ROLE)
		val := response[1:]
		
		cmdParts := strings.Fields(strings.ToUpper(command))
		if len(cmdParts) > 0 && cmdParts[0] == "KEYS" {
			// Format KEYS as a list
			keys := strings.Fields(val)
			if len(keys) == 0 {
				color.Yellow("(empty set)")
			} else {
				for i, k := range keys {
					fmt.Printf("%d) %s\n", i+1, color.CyanString(k))
				}
			}
		} else if len(cmdParts) > 0 && cmdParts[0] == "CLUSTER" {
			// Format CLUSTER output
			shards := strings.Split(val, ";")
			color.HiWhite("\nCluster Topology:")
			for _, s := range shards {
				if strings.TrimSpace(s) == "" {
					continue
				}
				parts := strings.Split(s, "Nodes:")
				header := strings.TrimSpace(parts[0])
				nodes := ""
				if len(parts) > 1 {
					nodes = strings.TrimSpace(parts[1])
				}

				if strings.Contains(header, "ALIVE") {
					fmt.Printf("  %s %s\n", color.GreenString("●"), header)
				} else {
					fmt.Printf("  %s %s\n", color.RedString("○"), header)
				}
				fmt.Printf("    Nodes: %s\n", color.CyanString(nodes))
			}
			fmt.Println()
		} else {
			color.Cyan(val)
		}
	} else if strings.HasPrefix(response, "-ERROR") {
		color.Red(response)
	} else {
		fmt.Println(response)
	}
}

func printHelp() {
	fmt.Println("Usage: carrotdb [options] [command]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  carrotdb                       # Start interactive REPL")
	fmt.Println("  carrotdb SET key value         # Set a value and exit")
	fmt.Println("  carrotdb GET key               # Get a value and exit")
	fmt.Println("  carrotdb -host 10.0.0.5 KEYS   # List keys on remote host")
}

func printInteractiveHelp() {
	color.HiWhite("\nAvailable Commands:")
	fmt.Printf("  %-15s %s\n", "SET <k> <v>", "Store a key-value pair")
	fmt.Printf("  %-15s %s\n", "GET <k>", "Retrieve value by key")
	fmt.Printf("  %-15s %s\n", "DELETE <k>", "Remove a key-value pair")
	fmt.Printf("  %-15s %s\n", "KEYS [prefix]", "List keys (optional prefix)")
	fmt.Printf("  %-15s %s\n", "CLUSTER", "Show cluster topology and health")
	fmt.Printf("  %-15s %s\n", "ROLE", "Show node's Raft role (Leader/Follower)")
	fmt.Printf("  %-15s %s\n", "COMPACT", "Trigger manual log compaction")
	fmt.Printf("  %-15s %s\n", "HELP", "Show this help message")
	fmt.Printf("  %-15s %s\n", "EXIT/QUIT", "Close the connection")
	fmt.Println()
}
