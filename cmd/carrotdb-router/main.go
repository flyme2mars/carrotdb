package main

import (
	"bufio"
	"carrotdb/pkg/sharding"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Router struct {
	addr        string
	ring        *sharding.HashRing
	shardPool   map[string]string // shardID -> leaderAddress
	shardStatus map[string]bool   // shardID -> isAlive
	conns       map[string]net.Conn // addr -> persistent connection
	mu          sync.RWMutex
}

func NewRouter(addr string) *Router {
	r := &Router{
		addr:        addr,
		ring:        sharding.NewHashRing(40),
		shardPool:   make(map[string]string),
		shardStatus: make(map[string]bool),
		conns:       make(map[string]net.Conn),
	}
	go r.startHealthCheck()
	return r
}

func (r *Router) startHealthCheck() {
	for {
		time.Sleep(5 * time.Second)
		r.mu.Lock()
		for id, addr := range r.shardPool {
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				r.shardStatus[id] = false
			} else {
				r.shardStatus[id] = true
				conn.Close()
			}
		}
		r.mu.Unlock()
	}
}

func (r *Router) AddShard(id string, leaderAddr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ring.AddShard(id)
	r.shardPool[id] = leaderAddr
	r.shardStatus[id] = true
}

func (r *Router) Start() error {
	// Start HTTP API
	go r.startHTTP(":8080")

	listener, err := net.Listen("tcp", r.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("🥕 Carrot-Router listening on %s (TCP) and :8080 (HTTP)", r.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		go r.handleClient(conn)
	}
}

func (r *Router) startHTTP(addr string) {
	http.HandleFunc("/api/status", r.handleStatus)
	http.Handle("/", http.FileServer(http.Dir("cmd/carrotdb-router/static")))
	log.Printf("📊 Dashboard API ready at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func (r *Router) handleStatus(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := struct {
		Shards map[string]string `json:"shards"`
		Status map[string]bool   `json:"status"`
	}{
		Shards: r.shardPool,
		Status: r.shardStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (r *Router) handleClient(clientConn net.Conn) {
	defer clientConn.Close()
	scanner := bufio.NewScanner(clientConn)

	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) < 2 {
			fmt.Fprintln(clientConn, "-ERROR: Usage: <COMMAND> <key> [value]")
			continue
		}

		key := parts[1]

		// 1. Find the correct shard for this key
		shardID := r.ring.GetShard(key)
		r.mu.RLock()
		shardAddr, ok := r.shardPool[shardID]
		r.mu.RUnlock()

		if !ok {
			fmt.Fprintf(clientConn, "-ERROR: shard %s not found in pool\r\n", shardID)
			continue
		}

		// 2. Forward the request to the correct Shard Leader
		response, err := r.forwardToShard(shardAddr, input)
		if err != nil {
			fmt.Fprintf(clientConn, "-ERROR: shard error: %v\r\n", err)
		} else {
			fmt.Fprint(clientConn, response)
		}
	}
}

func (r *Router) getShardConn(addr string) (net.Conn, error) {
	r.mu.RLock()
	conn, ok := r.conns[addr]
	r.mu.RUnlock()

	if ok {
		return conn, nil
	}

	// Create new connection if none exists
	newConn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.conns[addr] = newConn
	r.mu.Unlock()

	return newConn, nil
}

func (r *Router) forwardToShard(addr string, command string) (string, error) {
	conn, err := r.getShardConn(addr)
	if err != nil {
		return "", err
	}

	// Try to send command
	fmt.Fprintln(conn, command)
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')

	if err != nil {
		// Connection might be dead, try to reconnect once
		log.Printf("Connection to %s lost, retrying...", addr)
		conn.Close()
		
		r.mu.Lock()
		delete(r.conns, addr)
		r.mu.Unlock()

		newConn, err := r.getShardConn(addr)
		if err != nil {
			return "", err
		}
		fmt.Fprintln(newConn, command)
		return bufio.NewReader(newConn).ReadString('\n')
	}

	return resp, nil
}

func main() {
	router := NewRouter(":8000")

	// Static configuration for Phase 5 testing:
	// Shard 1 (Node 1)
	router.AddShard("shard1", "localhost:6379")
	// Shard 2 (Node 2)
	router.AddShard("shard2", "localhost:6380")

	log.Println("Starting CarrotDB Router with 2 shards...")
	if err := router.Start(); err != nil {
		log.Fatal(err)
	}
}
