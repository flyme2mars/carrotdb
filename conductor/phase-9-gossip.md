# CarrotDB Phase 9: Auto-Discovery & Dynamic Cluster Membership

The goal of this phase is to eliminate hardcoded IP addresses. CarrotDB nodes will automatically find each other and the Router will automatically discover the cluster topology using a Gossip Protocol.

## 🏗 Architecture

### 1. Gossip Protocol (Memberlist)
Every CarrotDB node (Server and Router) will join a **Gossip Network** using `hashicorp/memberlist`.
- **Discovery:** Nodes only need to know ONE "Seed" address to join the entire cluster.
- **Metadata:** Servers will broadcast their metadata:
    - `ShardID`: Which data partition they belong to.
    - `APIAddr`: Their client-facing TCP port.
    - `RaftAddr`: Their internal consensus port.

### 2. Router Auto-Discovery
The `Carrot-Router` will no longer have a hardcoded `AddShard` list.
- It will listen to Gossip "Join" events.
- When a new node joins, the Router reads its metadata and automatically adds it to the correct Shard in the Hash Ring.
- If a node leaves (or crashes), Gossip will detect it, and the Router will remove it from the pool.

### 3. Dynamic Rebalancing (Foundation)
When the Hash Ring changes (e.g., Shard 3 is added), the Router will immediately start sending new traffic to the new shard.
- We will implement a `MIGRATE` command in the Server protocol.
- Existing shards will identify keys that now belong to the new shard and "push" them over.

## 📂 Plan
1. **Integrate `hashicorp/memberlist`**: Add gossip support to both Server and Router.
2. **Implement Node Metadata**: Define the JSON structure for node identity.
3. **Update Router**: Replace static shard configuration with a dynamic listener that updates `shardPool` and `ring` in real-time.
4. **Dashboard Update**: Show "Discovered Nodes" on the dashboard.
5. **Rebalancing Logic**: Implement a basic key-migration background task that runs when the ring topology changes.

## 🚀 Final Goal (v0.9.0)
A truly "Zero-Config" database where you just start servers and they form a cluster automatically.
