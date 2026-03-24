package router

import (
	"bufio"
	"carrotdb/internal/server"
	"carrotdb/pkg/sharding"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

type Router struct {
	addr        string
	ring        *sharding.HashRing
	shardPool   map[string][]string // shardID -> list of all node addresses
	lastLeader  map[string]string   // shardID -> last known leader address
	shardStatus map[string]bool     // shardID -> isAlive (at least one node)
	conns       map[string]net.Conn // addr -> persistent connection
	mu          sync.RWMutex
	gossip      *memberlist.Memberlist
}

func NewRouter(addr string, gossip *memberlist.Memberlist) *Router {
	r := &Router{
		addr:        addr,
		ring:        sharding.NewHashRing(40),
		shardPool:   make(map[string][]string),
		lastLeader:  make(map[string]string),
		shardStatus: make(map[string]bool),
		conns:       make(map[string]net.Conn),
		gossip:      gossip,
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

func (r *Router) Start(httpAddr string, staticDir string) error {
	// Start HTTP API for Dashboard
	go r.startHTTP(httpAddr, staticDir)

	// Update Ring based on current gossip members immediately
	r.updateFromGossip()
	// Start background watcher for gossip changes
	go r.watchGossip()

	listener, err := net.Listen("tcp", r.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("📡 Internal Router listening on %s (TCP) and %s (HTTP)", r.addr, httpAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go r.handleClient(conn)
	}
}

func (r *Router) watchGossip() {
	for {
		time.Sleep(5 * time.Second)
		r.updateFromGossip()
	}
}

func (r *Router) updateFromGossip() {
	for _, member := range r.gossip.Members() {
		if len(member.Meta) == 0 {
			continue
		}
		var meta server.NodeMetadata
		if err := json.Unmarshal(member.Meta, &meta); err == nil && meta.ShardID != "" {
			apiAddr := meta.APIAddr
			if strings.HasPrefix(apiAddr, ":") {
				apiAddr = "127.0.0.1" + apiAddr
			}
			r.addNode(meta.ShardID, apiAddr)
		}
	}
}

func (r *Router) addNode(shardID string, addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes, exists := r.shardPool[shardID]
	if !exists {
		r.ring.AddShard(shardID)
		r.shardStatus[shardID] = true
	}

	for _, n := range nodes {
		if n == addr {
			return
		}
	}

	r.shardPool[shardID] = append(nodes, addr)
	if r.lastLeader[shardID] == "" {
		r.lastLeader[shardID] = addr
	}
}

func (r *Router) startHTTP(addr string, staticDir string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", r.handleStatus)
	if staticDir == "" {
		staticDir = "pkg/dashboard"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	log.Printf("📊 Dashboard available at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("HTTP server error: %v", err)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

		command := strings.ToUpper(parts[0])
		key := parts[1]

		// Handle non-sharded commands (ROLE, COMPACT, etc)
		if command == "ROLE" || command == "COMPACT" {
			found := false
			r.mu.RLock()
			shardIDs := make([]string, 0, len(r.shardPool))
			for id := range r.shardPool {
				shardIDs = append(shardIDs, id)
			}
			r.mu.RUnlock()

			for _, shardID := range shardIDs {
				response, err := r.forwardToShard(shardID, input)
				if err == nil {
					fmt.Fprint(clientConn, response)
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintln(clientConn, "-ERROR: no shards available")
			}
			continue
		}

		shardID := r.ring.GetShard(key)
		if shardID == "" {
			fmt.Fprintf(clientConn, "-ERROR: cluster not ready, discovering nodes...\r\n")
			continue
		}

		response, err := r.forwardToShard(shardID, input)
		if err != nil {
			fmt.Fprintf(clientConn, "-ERROR: %v\r\n", err)
		} else {
			fmt.Fprint(clientConn, response)
		}
	}
}

func (r *Router) forwardToShard(shardID string, command string) (string, error) {
	r.mu.RLock()
	addr := r.lastLeader[shardID]
	r.mu.RUnlock()

	resp, err := r.tryForward(addr, command)
	if err != nil || strings.Contains(resp, "not a Leader") {
		newAddr, err := r.findLeader(shardID)
		if err != nil {
			return "", err
		}
		return r.tryForward(newAddr, command)
	}

	return resp, nil
}

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

func (r *Router) tryForward(addr string, command string) (string, error) {
	r.mu.RLock()
	conn, ok := r.conns[addr]
	r.mu.RUnlock()

	if !ok {
		var err error
		conn, err = net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			return "", err
		}
		r.mu.Lock()
		r.conns[addr] = conn
		r.mu.Unlock()
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
