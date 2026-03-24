# CarrotDB: Architecture & Implementation Plan

CarrotDB is an educational, highly scalable Key-Value database written in Go. It is designed to be built in phases, starting from a simple in-memory store and evolving into a distributed, disk-backed database.

## Chosen Technologies
*   **Database Paradigm:** Key-Value Store
*   **Language:** Go (Golang)
*   **Interface:** Custom TCP/Text Protocol

## Roadmap

### Phase 1: The Core (In-Memory + Durability)
*Goal: Understand basic database operations and crash recovery.*
*   **Data Structure:** Implement a thread-safe in-memory key-value store using Go maps and `sync.RWMutex`.
*   **Operations:** `Put(key, value)`, `Get(key)`, `Delete(key)`.
*   **Durability:** Implement an Append-Only File (AOF) / Write-Ahead Log (WAL). Every write operation is appended to a file before confirming success. On startup, the database replays this log to rebuild its state.

### Phase 2: The Storage Engine (Bitcask Model)
*Goal: Break the RAM limit. Store datasets larger than available memory.*
*   **Architecture:** Adopt a Bitcask-like design (used in Riak). 
*   **Memory:** Keep only keys and their disk file offsets in memory (`map[string]Location`).
*   **Disk:** Read values directly from disk using the offset. Write speed remains fast because all writes are sequential appends.
*   **Compaction:** Build a background garbage collection process to merge old log files and remove stale/deleted keys to save disk space.

### Phase 3: The Network Layer
*Goal: Allow external applications to communicate with CarrotDB.*
*   **Server:** Build a custom TCP server using Go's `net` package.
*   **Protocol:** Design a simple text-based protocol (e.g., `SET mykey myvalue\r\n`, `GET mykey\r\n`).
*   **Concurrency:** Handle thousands of concurrent client connections efficiently using Go routines.

### Phase 4: Scaling & Distribution
*Goal: Make the database distributed, highly available, and fault-tolerant.*
*   **Sharding:** Implement Consistent Hashing to distribute data evenly across multiple CarrotDB nodes.
*   **Replication:** Integrate the Raft consensus algorithm to replicate data and elect leader nodes automatically.

### Phase 6: Web Dashboard
*Goal: Provide a visual interface for cluster monitoring.*
*   **Router API:** Small HTTP server inside the router to expose cluster state.
*   **Dashboard:** Monochromatic, minimal web UI to see shard health and stats.

### Phase 7: Performance Optimization
*Goal: Reduce latency and increase throughput.*
*   **Connection Pooling:** Persistent TCP connections from Router to Shards to avoid handshake overhead.
*   **Buffered I/O:** Using `bufio.Writer` in the Engine to batch disk writes and reduce syscalls.

### Phase 8: Smart Leader Discovery
*Goal: Automatically handle Raft leader changes.*
*   **Leader Probing:** Router queries nodes using the `ROLE` command to find the current leader.
*   **Automatic Failover:** If a leader fails, the Router automatically discovers the new leader without client-side errors.

### Phase 9: Auto-Discovery & Dynamic Rebalancing
*Goal: Remove hardcoded config and enable horizontal growth.*
*   **Gossip (Memberlist):** Nodes automatically discover each other via a gossip protocol.
*   **Dynamic Routing:** The Router automatically updates the Hash Ring when nodes join or leave.
*   **Automatic Rebalancing:** Servers detect shard changes and automatically migrate data to the new owners.

### Phase 10: Python SDK & Namespacing
*Goal: Multi-database support and cross-language usability.*
*   **Logical Namespacing:** SDK automatically prefixes keys with a database name to provide isolation.
*   **Python SDK:** A native Python library for easy integration into web apps.
*   **Prefix Searching:** Added `KEYS` command to the server protocol for namespace management.

### Phase 11: Symmetric Node Architecture
*Goal: Simplify minimal use-cases and remove single points of failure.*
*   **Embedded Router:** The traffic-routing logic is now built into every `carrotdb-server`.
*   **Unified Binary:** Removed the standalone router. Users now only need to run one type of program.
*   **Elastic Scaling:** Any node can act as an entry point. The cluster grows organically as nodes join.

---
*Status: Phase 11 (Symmetric) completed. CarrotDB is now a unified, modular distributed system.*