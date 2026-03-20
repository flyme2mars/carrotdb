# 🥕 CarrotDB

A high-performance, scalable, and educational Key-Value database written in Go.

CarrotDB is designed to be simple to understand but powerful enough to handle massive datasets. It uses a **Log-Structured Storage Engine** (Bitcask-inspired) to ensure extreme write speeds and crash resilience.

## ✨ Features (v0.5.0)
- **Horizontal Sharding:** Distribute data across multiple clusters (Shards) using **Consistent Hashing**.
- **Carrot-Router:** A smart proxy that automatically routes requests to the correct shard.
- **Distributed Replication:** Uses the **Raft Consensus Algorithm** for high availability within shards.
- **Scalable Index:** Only keys are kept in RAM, allowing for datasets larger than memory.

## 🚀 Quick Start

### 1. Installation
Download the binaries from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

### 2. Running a Sharded Cluster (Local Test)

**Start Shard 1 (Node 1):**
```bash
./carrotdb-server --id node1 --addr :6379 --raft :7000
```

**Start Shard 2 (Node 2):**
```bash
./carrotdb-server --id node2 --addr :6380 --raft :7001
```

**Start the Carrot-Router:**
```bash
./carrotdb-router
```

### 3. Usage
Connect your CLI to the **Router** (port 8000):
```bash
./carrotdb
> SET key1 value1
+OK
> SET key2 value2
+OK
```
The Router will automatically store `key1` on Shard 1 and `key2` on Shard 2!

## 🗺 Roadmap
- [x] **Phase 1:** In-Memory Store + Append-Only Log
- [x] **Phase 2:** Bitcask Storage Engine
- [x] **Phase 3:** TCP Networking & Custom Protocol
- [x] **Phase 4:** Distributed Consensus (Raft)
- [x] **Phase 5:** Horizontal Sharding & Consistent Hashing
- [ ] **Phase 6:** Dynamic Rebalancing & Auto-Discovery

## 📄 License
MIT
