# 🥕 CarrotDB

A high-performance, scalable, and educational Key-Value database written in Go.

CarrotDB is designed to be simple to understand but powerful enough to handle massive datasets. It uses a **Log-Structured Storage Engine** (Bitcask-inspired) to ensure extreme write speeds and crash resilience.

## ✨ Features (v0.4.0)
- **Distributed Replication:** Uses the **Raft Consensus Algorithm** to replicate data across multiple nodes.
- **High Availability:** Automatically elects a new leader if the current one fails.
- **Fast Writes:** Bitcask storage engine ensures extreme write speeds and disk resilience.
- **Scalable Index:** Only keys are kept in RAM, allowing for datasets larger than memory.
- **Crash Recovery:** Replays the write-ahead log (WAL) and Raft logs on startup.

## 🚀 Quick Start

### 1. Installation
Download the binary for your operating system from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

### 2. Running a Cluster (Local Test)

**Start Node 1 (Leader):**
```bash
./carrotdb-server --id node1 --addr :6379 --raft :7000
```

**Start Node 2 (Follower):**
```bash
./carrotdb-server --id node2 --addr :6380 --raft :7001 --join localhost:6379
```

**Start Node 3 (Follower):**
```bash
./carrotdb-server --id node3 --addr :6381 --raft :7002 --join localhost:6379
```

### 3. Usage
Connect to the **Leader** (node1) to write data:
```bash
./carrotdb
> SET key value
+OK
```

Connect to any **Follower** (node2 or node3) to read data:
```bash
# Connect to :6380
> GET key
+value
```

## 🗺 Roadmap
- [x] **Phase 1:** In-Memory Store + Append-Only Log
- [x] **Phase 2:** Bitcask Storage Engine
- [x] **Phase 3:** TCP Networking & Custom Protocol
- [x] **Phase 4:** Distributed Consensus (Raft)
- [ ] **Phase 5:** Horizontal Sharding & Consistent Hashing

## 📄 License
MIT
