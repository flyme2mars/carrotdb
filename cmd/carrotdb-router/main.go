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
	shardPool   map[string][]string // shardID -> list of all node addresses
	lastLeader  map[string]string   // shardID -> last known leader address
	shardStatus map[string]bool     // shardID -> isAlive (at least one node)
	conns       map[string]net.Conn // addr -> persistent connection
	mu          sync.RWMutex
}

func NewRouter(addr string) *Router {
	r := &Router{
		addr:        addr,
		ring:        sharding.NewHashRing(40),
		shardPool:   make(map[string][]string),
		lastLeader:  make(map[string]string),
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
		for id, addrs := range r.shardPool {
			anyAlive := false
			for _, addr := range addrs {
				conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
				if err == nil {
					anyAlive = true
					conn.Close()
					break
				}
			}
			r.shardStatus[id] = anyAlive
		}
		r.mu.Unlock()
	}
}

func (r *Router) AddShard(id string, nodes []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ring.AddShard(id)
	r.shardPool[id] = nodes
	if len(nodes) > 0 {
		r.lastLeader[id] = nodes[0]
	}
	r.shardStatus[id] = true
}

func (r *Router) getShardConn(addr string) (net.Conn, error) {
	r.mu.RLock()
	conn, ok := r.conns[addr]
	r.mu.RUnlock()

	if ok {
		return conn, nil
	}

	newConn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.conns[addr] = newConn
	r.mu.Unlock()

	return newConn, nil
}

// findLeader probes all nodes in a shard to find the current Raft leader.
func (r *Router) findLeader(shardID string) (string, error) {
	r.mu.RLock()
	addrs := r.shardPool[shardID]
	r.mu.RUnlock()

	for _, addr := range addrs {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			continue
		}
		
		fmt.Fprintln(conn, "ROLE")
		resp, err := bufio.NewReader(conn).ReadString('\n')
		conn.Close()

		if err == nil && strings.Contains(resp, "Leader") {
			r.mu.Lock()
			r.lastLeader[shardID] = addr
			r.mu.Unlock()
			return addr, nil
		}
	}

	return "", fmt.Errorf("no leader found for shard %s", shardID)
}

func (r *Router) forwardToShard(shardID string, command string) (string, error) {
	r.mu.RLock()
	addr := r.lastLeader[shardID]
	r.mu.RUnlock()

	resp, err := r.tryForward(addr, command)
	if err != nil || strings.Contains(resp, "not a Leader") {
		// Leader might have changed or node is down
		log.Printf("Leader for %s changed or unreachable, discovering...", shardID)
		newAddr, err := r.findLeader(shardID)
		if err != nil {
			return "", err
		}
		return r.tryForward(newAddr, command)
	}

	return resp, nil
}

func (r *Router) tryForward(addr string, command string) (string, error) {
	conn, err := r.getShardConn(addr)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(conn, command)
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')

	if err != nil {
		conn.Close()
		r.mu.Lock()
		delete(r.conns, addr)
		r.mu.Unlock()
		return "", err
	}

	return resp, nil
}

func (r *Router) Start() error {
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
			continue
		}
		go r.handleClient(conn)
	}
}

func (r *Router) startHTTP(addr string) {
	http.HandleFunc("/api/status", r.handleStatus)
	http.Handle("/", http.FileServer(http.Dir("cmd/carrotdb-router/static")))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func (r *Router) handleStatus(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := struct {
		Shards map[string][]string `json:"shards"`
		Status map[string]bool     `json:"status"`
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
		shardID := r.ring.GetShard(key)

		response, err := r.forwardToShard(shardID, input)
		if err != nil {
			fmt.Fprintf(clientConn, "-ERROR: %v\r\n", err)
		} else {
			fmt.Fprint(clientConn, response)
		}
	}
}

func main() {
	router := NewRouter(":8000")

	// Intelligent Shard Configuration:
	// We tell the router about ALL nodes in each shard cluster.
	router.AddShard("shard1", []string{"localhost:6379", "localhost:6380", "localhost:6381"})
	router.AddShard("shard2", []string{"localhost:6382", "localhost:6383", "localhost:6384"})

	log.Println("Starting Smart CarrotDB Router...")
	if err := router.Start(); err != nil {
		log.Fatal(err)
	}
}
