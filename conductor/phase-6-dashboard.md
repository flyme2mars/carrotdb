# CarrotDB Phase 6: Web Dashboard & Cluster Monitoring

The goal of this phase is to provide a **Visual Interface** to monitor the CarrotDB cluster. Instead of looking at logs, users can see the status of every shard, node, and the distribution of data across the Hash Ring.

## 🏗 Architecture
CarrotDB will add a small **HTTP API** to the `Carrot-Router` and a single-page **Web Dashboard**.

### 1. The Router API (REST)
The Router will expose its internal state via HTTP:
- `GET /status`: Returns a JSON object with:
    - List of Shards and their addresses.
    - Status of each Shard (Active/Inactive).
    - The Hash Ring structure (Virtual Nodes).
    - Statistics (Total keys, requests per second).

### 2. The Dashboard (Frontend)
A simple, built-in HTML/JavaScript dashboard.
- **Node Map:** A visual list of all servers and their roles (Leader/Follower).
- **Health Indicators:** Red/Green lights showing which servers are online.
- **Ring Visualization:** A circular map showing how data is distributed across the shards.

### 3. Automatic Health Checks
The Router will periodically "Ping" the underlying shards to check if they are still alive. If a shard goes down, the Router will mark it as "Inactive" on the dashboard.

## 📂 Plan
1. **Add HTTP Server to Router:** Integrate `net/http` into `cmd/carrotdb-router/main.go`.
2. **Implement `/api/status`:** Create a JSON response that serializes the Router's `shardPool` and `ring`.
3. **Embed Frontend Assets:** Use Go's `embed` package to bundle the HTML/CSS/JS into the binary.
4. **Health Checker:** Implement a background goroutine in the Router that pings shards every 5 seconds.
5. **UI Development:** Create a professional-looking dashboard with real-time updates (using `fetch` API).

## 🚀 Final Goal (v0.6.0)
A professional, observable database cluster that you can manage from your browser!
