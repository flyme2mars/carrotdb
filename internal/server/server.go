package server

import (
	"bufio"
	"carrotdb/internal/engine"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// Server represents a CarrotDB server with Raft support.
type Server struct {
	addr   string
	engine *engine.Engine
	raft   *raft.Raft
}

// NewServer creates a new instance of the Server with Raft initialized.
func NewServer(addr string, raftAddr string, nodeID string, engine *engine.Engine) (*Server, error) {
	// Initialize Raft
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// Setup FSM
	fsm := NewFSM(engine)

	// Setup Raft Storage
	dataDir := filepath.Join("data", nodeID)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.db"))
	if err != nil {
		return nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.db"))
	if err != nil {
		return nil, err
	}
	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Setup Transport
	raftAddrStr := raftAddr
	if strings.HasPrefix(raftAddrStr, ":") {
		raftAddrStr = "127.0.0.1" + raftAddrStr
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddrStr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(raftAddrStr, tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create Raft node
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	// Bootstrap the cluster (if it's the first node)
	hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		return nil, err
	}

	if !hasState {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(configuration)
	}

	return &Server{
		addr:   addr,
		engine: engine,
		raft:   r,
	}, nil
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

// handleConnection reads commands from a connection and executes them.
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
				if s.raft.State() != raft.Leader {
					fmt.Fprintf(conn, "-ERROR: Node is not a Leader. Current state: %s\r\n", s.raft.State())
					continue
				}

				key := parts[1]
				value := strings.Join(parts[2:], " ")
				cmd := Command{Op: "SET", Key: key, Value: value}
				data, _ := json.Marshal(cmd)

				future := s.raft.Apply(data, 10*time.Second)
				if err := future.Error(); err != nil {
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
				if s.raft.State() != raft.Leader {
					fmt.Fprintf(conn, "-ERROR: Node is not a Leader. Current state: %s\r\n", s.raft.State())
					continue
				}

				key := parts[1]
				cmd := Command{Op: "DELETE", Key: key}
				data, _ := json.Marshal(cmd)

				future := s.raft.Apply(data, 10*time.Second)
				if err := future.Error(); err != nil {
					fmt.Fprintf(conn, "-ERROR: %v\r\n", err)
				} else {
					fmt.Fprintln(conn, "+OK")
				}
			}

		case "JOIN":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "-ERROR: Usage: JOIN <node_id> <raft_addr>")
			} else {
				if s.raft.State() != raft.Leader {
					fmt.Fprintf(conn, "-ERROR: Only Leader can accept JOIN. Current state: %s\r\n", s.raft.State())
					continue
				}
				nodeID := parts[1]
				raftAddr := parts[2]
				future := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(raftAddr), 0, 0)
				if err := future.Error(); err != nil {
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
