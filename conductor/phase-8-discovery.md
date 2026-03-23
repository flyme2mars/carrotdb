# CarrotDB Phase 8: Smart Leader Discovery

The goal of this phase is to make the **Carrot-Router** intelligent enough to find the Leader of a shard automatically. This ensures that users never see a "Not a Leader" error, even if the cluster holds a new election.

## 🏗 Architecture

### 1. Multi-Node Shard Configuration
The Router will now be aware of all nodes in a shard, not just one.
```go
shardPool map[string][]string // shard1 -> ["node1:6379", "node2:6380", "node3:6381"]
```

### 2. Leader Probing Logic
When the Router receives a command for a shard:
1.  **Check Cache:** It tries the `lastKnownLeader` for that shard.
2.  **Probe on Failure:** If the connection fails or the node returns a "Not a Leader" error:
    -   The Router iterates through all known nodes for that shard.
    -   It sends a `LEADER?` probe to each node.
3.  **Update Cache:** Once a node confirms "I am Leader", the Router caches that address and completes the user's request.

### 3. Server-Side Support
We'll add a simple `ROLE` command to the CarrotDB protocol:
-   **Request:** `ROLE\n`
-   **Response:** `+LEADER\n` or `+FOLLOWER\n`

## 📂 Plan
1.  **Update `internal/server/server.go`**: Add the `ROLE` command.
2.  **Update `cmd/carrotdb-router/main.go`**:
    -   Update `Router` struct to store multiple addresses per shard.
    -   Update `AddShard` to accept a list of addresses.
    -   Implement `findLeader(shardID)` logic that probes nodes.
    -   Update `forwardToShard` to use the discovery logic.
3.  **Refactor `main()` in Router**: Update the static configuration to include multiple nodes.
4.  **Verification**: Start 3 nodes, manually kill the leader, and verify the Router automatically finds the new leader without client errors.

## 🚀 Final Goal (v0.8.0)
A truly resilient database cluster where the proxy layer handles all complexity of distributed elections.
