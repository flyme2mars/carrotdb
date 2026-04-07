# 🥕 CarrotDB (v1.0.0 Stable)

**A high-performance, sharded, and self-healing Key-Value database written in Go.**

CarrotDB is an industrial-grade distributed database designed for extreme scalability and crash resilience. It combines the speed of **Bitcask** storage with the safety of **Raft** consensus and the horizontal power of **Consistent Hashing**.

**Read the [Architecture Guide](ARCHITECTURE.md) for a deep dive into the theory.**

---

## ✨ Features

*   **Unified Architecture:** Every node acts as a gateway. No separate routers needed.
*   **Self-Healing:** Nodes automatically discover each other using a **Gossip Protocol**.
*   **Horizontal Sharding:** Distribute data across unlimited shards with zero configuration.
*   **High Availability:** Built-in **Raft Consensus** ensures zero data loss during node failures.
*   **Snapshotting:** Automatic Raft log compaction for infinite durability without disk bloat.
*   **Multi-Tenancy:** SDK-level logical namespacing provides isolated "databases" on a single cluster.
*   **V2 Dashboard:** Professional, minimalist monochromatic UI for real-time cluster monitoring.
*   **Cluster Topology:** CLI-level insight into shard health and node distribution.
*   **Trace Mode:** Educational "--trace" flag to visualize the internal request lifecycle across all layers.
*   **Lab Bench:** Quantitative benchmarking tool to measure and explain distributed system trade-offs.
*   **Cloud Native:** First-class Docker and Kubernetes support with environment-based configuration.

---

## 🚀 Getting Started

### 1. Installation

**Using Docker (Recommended):**
```bash
docker pull ghcr.io/flyme2mars/carrotdb:latest
```

**Using Go:**
```bash
go install github.com/flyme2mars/carrotdb/cmd/carrotdb-server@latest
go install github.com/flyme2mars/carrotdb/cmd/carrotdb@latest
```

---

## 🐳 Docker Deployment

### Single Node
```bash
docker run -d \
  --name carrotdb \
  -p 8000:8000 -p 8080:8080 \
  -v carrot_data:/home/carrot/data \
  ghcr.io/flyme2mars/carrotdb:latest
```

### 3-Node Cluster (Docker Compose)
We provide a ready-to-use `docker-compose.yaml`. Simply run:
```bash
docker-compose up -d
```
This spins up a multi-shard, high-availability cluster with persistent storage and auto-discovery.

---

## 🎓 Educational: Trace Mode
CarrotDB is built for learning. Run any node with the `--trace` flag to see the "hidden" lifecycle of every request:
```bash
./carrotdb-server --trace
```
**You will see color-coded logs for:**
*   **Blue (Router):** How keys are hashed and mapped to shards.
*   **Cyan (Server/Raft):** When commands are proposed and committed to the cluster.
*   **Magenta (Engine):** Exactly where and when data is written to the physical append-only log.

---

## 🧪 Educational: Lab Bench
The **Lab Bench** (`carrotdb-bench`) is a scientific instrument for CarrotDB. It runs standardized workloads and generates an educational report explaining the **"Why"** behind the performance.

### Run an Experiment
```bash
# Start a benchmark against a running cluster
./carrotdb-bench -port 8000 -duration 10s
```

**It measures and explains:**
*   **Sharding Impact:** How throughput increases as you add more shards.
*   **Fan-out Latency:** Why global commands like `KEYS` get slower as the cluster grows.
*   **Concurrency:** How multiple workers interact with the Raft leader.

---

**From Binaries:**
Download the pre-compiled binaries for your OS from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

### 2. Basic Setup (Single Node)
To use CarrotDB as a simple local Key-Value store:
```bash
./carrotdb-server
```
*   **API:** Port 8000 (TCP)
*   **Storage:** Port 6379 (Internal)
*   **Dashboard:** [http://localhost:8080](http://localhost:8080)

### 3. Advanced Setup (Distributed Cluster)
CarrotDB scales horizontally by adding shards.

**Node 1 (The Seed):**
```bash
./carrotdb-server --id node1 --addr :6379 --raft :7000 --gossip-addr :9000
```

**Node 2 (Join Shard 1):**
```bash
./carrotdb-server --id node2 --addr :6380 --raft :7001 --gossip-addr :9001 --gossip-seed 127.0.0.1:9000
```

**Node 3 (New Shard - Horizontal Scaling):**
```bash
./carrotdb-server --id node3 --shard shard2 --addr :6381 --raft :7002 --gossip-addr :9002 --gossip-seed 127.0.0.1:9000
```

---

## 🔌 Client SDKs

### Python SDK
Install the official Python client:
```bash
cd sdk/python
pip install .
```

**Usage Example:**
```python
from carrotdb import Client

# Connect to any node in the cluster
db = Client(host="localhost", port=8000, database="my_project")

db.set("user:1", "Alice")
print(db.get("user:1")) # Outputs: Alice
```

---

## 📊 Monitoring
Every CarrotDB node hosts a built-in dashboard. Open **[http://localhost:8080](http://localhost:8080)** in your browser to view:
*   Real-time health of every node.
*   Shard distribution and roles (Leader/Follower).
*   Cluster-wide statistics and uptime.

---

## 📄 License
MIT License.
