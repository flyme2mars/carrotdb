package server

import (
	"bufio"
	"carrotdb/internal/engine"
	"carrotdb/pkg/sharding"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// NodeMetadata represents the information broadcast via Gossip.
type NodeMetadata struct {
	ShardID  string `json:"shard_id"`
	APIAddr  string `json:"api_addr"`
	RaftAddr string `json:"raft_addr"`
}

// Server represents a CarrotDB server with Raft and Gossip support.
type Server struct {
	addr    string
	shardID string
	engine  *engine.Engine
	raft    *raft.Raft
	gossip  *memberlist.Memberlist
	ring    *sharding.HashRing
	mu      sync.RWMutex
}

// NewServer creates a new instance of the Server with Raft and Gossip initialized.
func NewServer(addr string, raftAddr string, nodeID string, shardID string, engine *engine.Engine, gossipAddr string, gossipSeed string) (*Server, error) {
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

	// Initialize Gossip (Memberlist)
	mConfig := memberlist.DefaultLocalConfig()
	mConfig.Name = nodeID
	
	// SILENCE SLOP: Only log errors, not every ping/pong
	mConfig.LogOutput = io.Discard 
	
	// Normalize addresses for metadata
	fullAPIAddr := addr
	if strings.HasPrefix(fullAPIAddr, ":") {
		fullAPIAddr = "127.0.0.1" + fullAPIAddr
	}
	fullRaftAddr := raftAddr
	if strings.HasPrefix(fullRaftAddr, ":") {
		fullRaftAddr = "127.0.0.1" + fullRaftAddr
	}

	// Force IPv4 to avoid [::] vs 127.0.0.1 conflicts
	mConfig.BindAddr = "127.0.0.1"
	if gossipAddr != "" {
		_, portStr, _ := net.SplitHostPort(gossipAddr)
		mConfig.BindPort, _ = strconv.Atoi(portStr)
	}

	meta := NodeMetadata{
		ShardID:  shardID,
		APIAddr:  fullAPIAddr,
		RaftAddr: fullRaftAddr,
	}
	mConfig.Delegate = &gossipDelegate{meta: meta}

	m, err := memberlist.Create(mConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create memberlist: %w", err)
	}

	if gossipSeed != "" {
		_, err := m.Join([]string{gossipSeed})
		if err != nil {
			log.Printf("failed to join gossip cluster: %v", err)
		}
	}

	s := &Server{
		addr:    addr,
		shardID: shardID,
		engine:  engine,
		raft:    r,
		gossip:  m,
		ring:    sharding.NewHashRing(40),
	}

	go s.monitorRing()

	return s, nil
}

func (s *Server) monitorRing() {
	for {
		time.Sleep(10 * time.Second)
		s.updateRing()
		if s.raft.State() == raft.Leader {
			s.rebalance()
		}
	}
}

func (s *Server) rebalance() {
	s.mu.RLock()
	ring := s.ring
	s.mu.RUnlock()

	keys := s.engine.Keys()
	for _, key := range keys {
		targetShard := ring.GetShard(key)
		if targetShard != "" && targetShard != s.shardID {
			log.Printf("Rebalance: Key %s belongs to %s (current: %s). Migrating...", key, targetShard, s.shardID)
			s.migrateKey(key, targetShard)
		}
	}
}

func (s *Server) migrateKey(key string, targetShard string) {
	// 1. Get value
	val, err := s.engine.Get(key)
	if err != nil {
		return
	}

	// 2. Find target leader (via Gossip)
	var targetAddr string
	for _, member := range s.gossip.Members() {
		var meta NodeMetadata
		json.Unmarshal(member.Meta, &meta)
		if meta.ShardID == targetShard {
			targetAddr = meta.APIAddr
			break
		}
	}

	if targetAddr == "" {
		return
	}

	// 3. Send to target
	conn, err := net.DialTimeout("tcp", targetAddr, 2*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "SET %s %s\n", key, val)
	resp, _ := bufio.NewReader(conn).ReadString('\n')

	if strings.Contains(resp, "+OK") {
		// 4. Delete locally via Raft
		cmd := Command{Op: "DELETE", Key: key}
		data, _ := json.Marshal(cmd)
		s.raft.Apply(data, 5*time.Second)
	}
}

func (s *Server) updateRing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	newRing := sharding.NewHashRing(40)
	shards := make(map[string]bool)

	for _, member := range s.gossip.Members() {
		var meta NodeMetadata
		if err := json.Unmarshal(member.Meta, &meta); err != nil {
			continue
		}
		if meta.ShardID != "" && !shards[meta.ShardID] {
			newRing.AddShard(meta.ShardID)
			shards[meta.ShardID] = true
		}
	}
	s.ring = newRing
}

type gossipDelegate struct {
	meta NodeMetadata
}

func (d *gossipDelegate) NodeMeta(limit int) []byte {
	data, _ := json.Marshal(d.meta)
	return data
}

func (d *gossipDelegate) NotifyMsg([]byte)                           {}
func (d *gossipDelegate) GetBroadcasts(overhead, limit int) [][]byte { return nil }
func (d *gossipDelegate) LocalState(join bool) []byte                { return nil }
func (d *gossipDelegate) MergeRemoteState(buf []byte, join bool)     {}

// Gossip returns the memberlist instance.
func (s *Server) Gossip() *memberlist.Memberlist {
	return s.gossip
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

		case "ROLE":
			fmt.Fprintf(conn, "+%s\r\n", s.raft.State())

		case "KEYS":
			prefix := ""
			if len(parts) >= 2 {
				prefix = parts[1]
			}
			keys := s.engine.KeysWithPrefix(prefix)
			fmt.Fprintf(conn, "+%s\r\n", strings.Join(keys, " "))

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
