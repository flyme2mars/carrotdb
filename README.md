# 🥕 CarrotDB

A high-performance, scalable, and educational Key-Value database written in Go.

CarrotDB is designed to be simple to understand but powerful enough to handle massive datasets. It uses a **Log-Structured Storage Engine** (Bitcask-inspired) to ensure extreme write speeds and crash resilience.

## ✨ Pro Features
- **Unified Binary:** Every node is a gateway. No separate router needed for simple or complex clusters.
- **Self-Healing Cluster:** Nodes automatically discover each other using a **Gossip Protocol**.
- **Horizontal Sharding:** Distribute data across unlimited shards with zero configuration.
- **High Availability:** Built-in **Raft Consensus** ensures zero data loss during node failures.
- **V2 Dashboard:** Minimalist, monochromatic cluster grid for real-time monitoring.

## 🚀 Quick Start (One Command)

### 1. Installation
Download the latest binaries from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

### 2. Run a Single Node
```bash
./carrotdb-server
```
That's it! CarrotDB is now running its storage engine AND router.

### 3. Start a Cluster (Example)

**Node 1 (Seed):**
```bash
./carrotdb-server --id node1 --addr :6379 --gossip-addr :9000
```

**Node 2 (Join):**
```bash
./carrotdb-server --id node2 --addr :6380 --gossip-addr :9001 --gossip-seed 127.0.0.1:9000
```

**Connect to ANY node on port 8000!** Both nodes will automatically route your data to the correct shard.

### 4. Monitoring
Open your browser to **[http://localhost:8080](http://localhost:8080)** to view the cluster grid.

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
