# CarrotDB Phase 4: Scaling & Distribution (The Cluster)

The final goal of CarrotDB is to become a **Distributed Database**. Instead of one server, we will have a **Cluster** of servers that work together to store more data (Scaling) and survive failures (Availability).

## 🏗 Architecture
CarrotDB will use a **Leader-Follower** model based on the **Raft Consensus Algorithm**.

### 1. The Raft Cluster (Replication)
To ensure that we never lose data if one server crashes, we will have 3 or 5 nodes in a "Raft Cluster".
- **Leader:** All writes (`SET`, `DELETE`) go to the Leader.
- **Followers:** The Leader replicates its Append-Only Log to all Followers.
- **Agreement:** A write is only confirmed as "OK" once a majority (quorum) of nodes have safely written it to their own disk.
- **Auto-Election:** If the Leader crashes, the Followers will automatically detect it and elect a new Leader in milliseconds.

### 2. Consistent Hashing (Sharding)
To store more data than one machine can handle, we will use **Sharding**.
- The keyspace is mapped onto a "Hash Ring".
- Each group of nodes (a Raft Shard) is responsible for a segment of that ring.
- **Carrot-Router:** A thin proxy layer that knows the topology of the ring and redirects client requests to the correct shard.

### 3. Distributed Protocol
We need to update our protocol to handle internal cluster communication:
- `RAFT_JOIN`: A new node joins the cluster.
- `RAFT_VOTE`: Nodes voting for a new leader.
- `RAFT_APPEND`: Replicating the log from leader to follower.

## 📂 Plan
1. **Integrate `hashicorp/raft`:** Use the industry-standard Raft implementation for Go.
2. **Cluster Configuration:** Allow `carrotdb-server` to start in "Cluster Mode" with an ID and a list of peer addresses.
3. **Log Replication:** Hook our existing `Engine.Put()` and `Engine.Delete()` into the Raft FSM (Finite State Machine).
4. **Leader Forwarding:** If a client talks to a Follower, the server should either reject the write or automatically forward it to the Leader.
5. **Node Discovery:** Implement a simple mechanism for nodes to find each other (Gossip protocol or static config).
6. **Verification:** Start 3 servers, kill one, and verify that you can still `GET` your data from the remaining ones.

## 🚀 Final Goal (v1.0.0)
A fully fault-tolerant, horizontally scalable database that you can deploy on Kubernetes or a multi-server cluster.
