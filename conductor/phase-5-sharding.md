# CarrotDB Phase 5: Horizontal Sharding & Consistent Hashing

The goal of this phase is to achieve **Horizontal Scalability**. Instead of one Raft cluster, we will have multiple independent Raft clusters (Shards). This allows CarrotDB to store petabytes of data by simply adding more machines.

## 🏗 Architecture
CarrotDB will use **Consistent Hashing** to distribute keys across multiple Shards.

### 1. The Hash Ring
- We'll create a "Hash Ring" (a 0 to 2^32 circle).
- Each Raft Shard (a group of 3 nodes) is assigned multiple "Virtual Nodes" on the ring.
- **Why?** Virtual nodes ensure that data is distributed evenly even if some shards are larger than others.

### 2. The Carrot-Router (The Proxy)
To keep the database simple for users, they will talk to a **Router** instead of talking directly to the shards.
- The user sends `SET mykey myvalue` to the Router.
- The Router hashes `mykey`, finds the correct Shard on the Hash Ring, and forwards the request to that Shard's Leader.
- The user doesn't need to know which shard their data is on.

### 3. Distributed Topology
We need a way for the Router to know which nodes belong to which shards.
- **Shard Registry:** A simple configuration or a separate "Coordinator" node that tracks the current cluster topology.

## 📂 Plan
1. **Implement Consistent Hashing:** Create a `pkg/sharding` package that manages the Hash Ring.
2. **Build the Carrot-Router:** Create a new binary `cmd/carrotdb-router` that acts as a thin proxy.
3. **Multi-Shard Support:** Update `carrotdb-server` to identify which shard it belongs to (e.g., `--shard shard1`).
4. **Proxy Logic:** The Router will use a pool of TCP connections to talk to the underlying shards.
5. **Verification:** Start 2 separate shards (6 nodes total) and 1 Router. Verify that data is split correctly between the two shards.

## 🚀 Final Goal (v0.5.0)
A truly distributed, sharded database that is ready for industrial use cases!
