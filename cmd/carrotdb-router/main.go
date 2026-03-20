package main

import (
	"bufio"
	"carrotdb/pkg/sharding"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Router struct {
	addr      string
	ring      *sharding.HashRing
	shardPool map[string]string // shardID -> leaderAddress
	mu        sync.RWMutex
}

func NewRouter(addr string) *Router {
	return &Router{
		addr:      addr,
		ring:      sharding.NewHashRing(40),
		shardPool: make(map[string]string),
	}
}

func (r *Router) AddShard(id string, leaderAddr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ring.AddShard(id)
	r.shardPool[id] = leaderAddr
}

func (r *Router) Start() error {
	listener, err := net.Listen("tcp", r.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("🥕 Carrot-Router listening on %s", r.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		go r.handleClient(conn)
	}
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

		command := strings.ToUpper(parts[0])
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

func (r *Router) forwardToShard(addr string, command string) (string, error) {
	// In a production system, we would use a connection pool here
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	fmt.Fprintln(conn, command)
	reader := bufio.NewReader(conn)
	return reader.ReadString('\n')
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
