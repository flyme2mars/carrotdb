package main

import (
	"carrotdb/internal/engine"
	"carrotdb/internal/router"
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

// getEnv returns the value of an environment variable or a default value if not set.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	var (
		nodeID     = flag.String("id", getEnv("CARROT_ID", "node1"), "Node ID (Env: CARROT_ID)")
		shardID    = flag.String("shard", getEnv("CARROT_SHARD", "shard1"), "Shard ID (Env: CARROT_SHARD)")
		httpAddr   = flag.String("addr", getEnv("CARROT_ADDR", ":6379"), "Client API address (Env: CARROT_ADDR)")
		raftAddr   = flag.String("raft", getEnv("CARROT_RAFT", ":7000"), "Raft internal address (Env: CARROT_RAFT)")
		joinAddr   = flag.String("join", getEnv("CARROT_JOIN", ""), "Address of the leader to join (Env: CARROT_JOIN)")
		gossipAddr = flag.String("gossip-addr", getEnv("CARROT_GOSSIP_ADDR", ":9000"), "Gossip bind address (Env: CARROT_GOSSIP_ADDR)")
		gossipSeed = flag.String("gossip-seed", getEnv("CARROT_GOSSIP_SEED", ""), "Gossip seed address (Env: CARROT_GOSSIP_SEED)")
		routerAddr = flag.String("router-addr", getEnv("CARROT_ROUTER_ADDR", ":8000"), "Internal router TCP address (Env: CARROT_ROUTER_ADDR)")
		dashAddr   = flag.String("dashboard-addr", getEnv("CARROT_DASHBOARD_ADDR", ":8080"), "Dashboard HTTP address (Env: CARROT_DASHBOARD_ADDR)")
		staticDir  = flag.String("static-dir", getEnv("CARROT_STATIC_DIR", "static"), "Directory for dashboard static files (Env: CARROT_STATIC_DIR)")
		dataBase   = flag.String("data-dir", getEnv("CARROT_DATA_DIR", "data"), "Base directory for data (Env: CARROT_DATA_DIR)")
		trace      = flag.Bool("trace", false, "Enable educational protocol tracing")
	)
	flag.Parse()

	log.Printf("🥕 CarrotDB starting (Node: %s, Shard: %s, API: %s, Raft: %s, Gossip: %s)", *nodeID, *shardID, *httpAddr, *raftAddr, *gossipAddr)

	// Ensure the data directory exists for this node
	dataDir := filepath.Join(*dataBase, *nodeID)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize the storage engine for this node
	db, err := engine.NewEngine(filepath.Join(dataDir, "carrotdb.log"), *trace)
	if err != nil {
		log.Fatalf("failed to start engine: %v", err)
	}
	defer db.Close()

	// Initialize and start the Server with Raft and Gossip
	s, err := server.NewServer(*httpAddr, *raftAddr, *nodeID, *shardID, db, *gossipAddr, *gossipSeed, *trace)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// Start Internal Router
	r := router.NewRouter(*routerAddr, s.Gossip(), *trace)
	go func() {
		if err := r.Start(*dashAddr, *staticDir); err != nil {
			log.Printf("failed to start internal router: %v", err)
		}
	}()

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
