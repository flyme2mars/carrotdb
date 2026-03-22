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

---
*Status: Phase 7 (Performance) completed. CarrotDB is now optimized for high-throughput workloads.*