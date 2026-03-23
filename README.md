# 🥕 CarrotDB

A high-performance, scalable, and educational Key-Value database written in Go.

CarrotDB is designed to be simple to understand but powerful enough to handle massive datasets. It uses a **Log-Structured Storage Engine** (Bitcask-inspired) to ensure extreme write speeds and crash resilience.

## ✨ Features (v0.8.0)
- **Smart Leader Discovery:** The Router automatically finds the cluster Leader for every shard. No more "Not a Leader" errors!
- **High Performance:** Persistent **Connection Pooling** and **Buffered Writes** for maximum throughput.
- **Web Dashboard:** Monochromatic, minimal visual monitoring of cluster health.
- **Horizontal Sharding:** Distribute data across multiple clusters using **Consistent Hashing**.
- **Distributed Replication:** Uses the **Raft Consensus Algorithm** for high availability.

## 🚀 Quick Start

### 1. Installation
Download the binaries from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

### 2. Running a Sharded Cluster (Local Test)

**Start Shards:**
```bash
./carrotdb-server --id node1 --addr :6379 --raft :7000
./carrotdb-server --id node2 --addr :6380 --raft :7001
```

**Start the Carrot-Router:**
```bash
./carrotdb-router
```

### 3. Monitoring (Dashboard)
Open your browser and navigate to:
**[http://localhost:8080](http://localhost:8080)**

You will see a real-time list of all shards and their current health status (Online/Offline).

### 4. Usage

Connect your CLI to the **Router** (port 8000):
```bash
./carrotdb
> SET key1 value1
+OK
> SET key2 value2
+OK
```
The Router will automatically store `key1` on Shard 1 and `key2` on Shard 2!

## 📄 License
MIT
