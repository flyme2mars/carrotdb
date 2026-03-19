package server

import (
	"bufio"
	"carrotdb/internal/engine"
	"fmt"
	"log"
	"net"
	"strings"
)

// Server represents a CarrotDB server.
type Server struct {
	addr   string
	engine *engine.Engine
}

// NewServer creates a new instance of the Server.
func NewServer(addr string, engine *engine.Engine) *Server {
	return &Server{
		addr:   addr,
		engine: engine,
	}
}

// Start listens for incoming connections and handles them in separate goroutines.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	defer listener.Close()

	log.Printf("🥕 CarrotDB server listening on %s", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		// Handle connection in a new goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection reads commands from a connection and executes them on the engine.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Client connected: %s", conn.RemoteAddr())
	defer log.Printf("Client disconnected: %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "QUIT", "EXIT":
			fmt.Fprintln(conn, "+OK")
			return

		case "SET":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "-ERROR: Usage: SET <key> <value>")
			} else {
				key := parts[1]
				value := strings.Join(parts[2:], " ")
				if err := s.engine.Put(key, value); err != nil {
					fmt.Fprintf(conn, "-ERROR: %v\r\n", err)
				} else {
					fmt.Fprintln(conn, "+OK")
				}
			}

		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "-ERROR: Usage: GET <key>")
			} else {
				key := parts[1]
				val, err := s.engine.Get(key)
				if err != nil {
					fmt.Fprintf(conn, "-ERROR: %v\r\n", err)
				} else {
					fmt.Fprintf(conn, "+%s\r\n", val)
				}
			}

		case "DELETE":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "-ERROR: Usage: DELETE <key>")
			} else {
				key := parts[1]
				if err := s.engine.Delete(key); err != nil {
					fmt.Fprintf(conn, "-ERROR: %v\r\n", err)
				} else {
					fmt.Fprintln(conn, "+OK")
				}
			}

		case "COMPACT":
			if err := s.engine.Compact(); err != nil {
				fmt.Fprintf(conn, "-ERROR: %v\r\n", err)
			} else {
				fmt.Fprintln(conn, "+OK")
			}

		default:
			fmt.Fprintf(conn, "-ERROR: Unknown command: %s\r\n", command)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Connection error (%s): %v", conn.RemoteAddr(), err)
	}
}
