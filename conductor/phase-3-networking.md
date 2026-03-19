# CarrotDB Phase 3: The Network Layer

The goal of this phase is to turn CarrotDB from a local CLI tool into a **Database Server** that can accept multiple concurrent connections over TCP.

## 🏗 Architecture
CarrotDB will follow the **Multi-threaded TCP Server** model, which Go handles exceptionally well using goroutines.

### 1. TCP Server
- **Listener:** Uses `net.Listen` to wait for incoming connections on a port (e.g., `6379`).
- **Connection Handler:** Every new client gets its own **Goroutine**. This allows CarrotDB to handle thousands of users simultaneously without blocking.
- **Graceful Shutdown:** Ensures all connections are closed and data is synced to disk when the server stops.

### 2. CarrotDB Protocol (Text-based)
To keep it simple and debuggable for now, we will use a human-readable text protocol similar to Redis.

**Requests:**
- `SET <key> <value>\n`
- `GET <key>\n`
- `DELETE <key>\n`
- `COMPACT\n` (Manually trigger storage compaction)
- `QUIT\n`

**Responses:**
- `+OK\n` (Success)
- `+<value>\n` (Retrieved value)
- `-ERROR: <message>\n` (Error message)

### 3. Concurrency Safety
The `Engine` from Phase 2 already uses `sync.RWMutex`, which is perfect for this. It ensures that multiple network clients can safely read and write to the same data file.

## 📂 Plan
1. **Create `internal/server` package:** Implement the main loop and connection handling.
2. **Implement Protocol Parser:** A utility to read lines from the TCP connection and identify commands.
3. **Dispatch to Engine:** Connect the network commands to the `Put`, `Get`, `Delete`, and `Compact` methods of our storage engine.
4. **New Entry Point:** Create `cmd/carrotdb-server/main.go` to start the server.
5. **Update CLI (optional):** Refactor the current CLI to act as a "Client" that connects to the server instead of talking directly to the engine.
6. **Verification:** Test with multiple `telnet` or `netcat` sessions simultaneously.

## 🚀 Future (Phase 3.1)
- **Binary Protocol (RESP):** Migrate to a binary protocol for higher performance and compatibility with Redis clients.
- **Connection Pooling:** Optimize how clients stay connected.
