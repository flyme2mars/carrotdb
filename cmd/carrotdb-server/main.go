package main

import (
	"carrotdb/internal/engine"
	"carrotdb/internal/server"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var (
		nodeID     = flag.String("id", "node1", "Node ID")
		shardID    = flag.String("shard", "shard1", "Shard ID")
		httpAddr   = flag.String("addr", ":6379", "Client API address")
		raftAddr   = flag.String("raft", ":7000", "Raft internal address")
		joinAddr   = flag.String("join", "", "Address of the leader to join (Raft)")
		gossipAddr = flag.String("gossip-addr", ":9000", "Gossip bind address")
		gossipSeed = flag.String("gossip-seed", "", "Gossip seed address")
	)
	flag.Parse()

	log.Printf("🥕 CarrotDB starting (Node: %s, Shard: %s, API: %s, Raft: %s, Gossip: %s)", *nodeID, *shardID, *httpAddr, *raftAddr, *gossipAddr)

	// Ensure the data directory exists for this node
	dataDir := filepath.Join("data", *nodeID)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize the storage engine for this node
	db, err := engine.NewEngine(filepath.Join(dataDir, "carrotdb.log"))
	if err != nil {
		log.Fatalf("failed to start engine: %v", err)
	}
	defer db.Close()

	// Initialize and start the Server with Raft and Gossip
	s, err := server.NewServer(*httpAddr, *raftAddr, *nodeID, *shardID, db, *gossipAddr, *gossipSeed)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// If joinAddr is specified, try to join the cluster
	if *joinAddr != "" {
		go func() {
			// Wait a bit for the server to start
			time.Sleep(2 * time.Second)
			if err := joinCluster(*joinAddr, *nodeID, *raftAddr); err != nil {
				log.Printf("failed to join cluster: %v", err)
			}
		}()
	}

	if err := s.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func joinCluster(leaderAddr, nodeID, raftAddr string) error {
	conn, err := net.Dial("tcp", leaderAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send JOIN command to the leader
	command := fmt.Sprintf("JOIN %s %s\n", nodeID, raftAddr)
	if _, err := conn.Write([]byte(command)); err != nil {
		return err
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	response := strings.TrimSpace(string(buf[:n]))
	if response != "+OK" {
		return fmt.Errorf("leader rejected join: %s", response)
	}

	log.Printf("Successfully joined cluster at %s", leaderAddr)
	return nil
}
