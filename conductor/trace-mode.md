# Plan: Educational "Trace Mode" for CarrotDB

This plan introduces a "Trace Mode" to CarrotDB, allowing users and students to "see through" the database's internal operations as they happen.

## Objective
Expose the "hidden" lifecycle of every request (from Router to Raft to Engine) via color-coded console logs on the server.

## Changes

### 1. `cmd/carrotdb-server/main.go`
- Add a `-trace` flag (default `false`).
- Pass the `trace` boolean to the `Engine`, `Server`, and `Router` initializers.

### 2. `internal/engine/engine.go`
- Add a `Trace` boolean to the `Engine` struct.
- Add Magenta-colored logs in `Put` (log offset), `Get` (reading offset/size), and `Delete`.
- Add logs in `Compact` to show the merge progress.

### 3. `internal/server/server.go`
- Add a `Trace` boolean to the `Server` struct.
- Add Cyan-colored logs in `handleConnection` when a command is received.
- Add logs in `SET` and `DELETE` when a command is proposed to Raft.
- Add logs in `GET` when a local read is performed.

### 4. `internal/router/router.go`
- Add a `Trace` boolean to the `Router` struct.
- Add Blue-colored logs in `handleClient` when a key is hashed and mapped to a shard.
- Add logs in `forwardToShard` when a command is routed to a specific leader address.

## Implementation Steps

### Phase 1: Engine Tracing
1. Update `Engine` struct and `NewEngine`.
2. Add `[TRACE] Engine: ...` logs in core operations.

### Phase 2: Server Tracing
1. Update `Server` struct and `NewServer`.
2. Add `[TRACE] Server: ...` logs in command handling.

### Phase 3: Router Tracing
1. Update `Router` struct and `NewRouter`.
2. Add `[TRACE] Router: ...` logs in hashing and forwarding.

### Phase 4: Final Integration
1. Update `cmd/carrotdb-server/main.go` to link the `-trace` flag to all components.

## Verification & Testing
1. Run server with `--trace`: `./carrotdb-server --trace`.
2. Run a command from CLI: `./carrotdb SET user:1 Alice`.
3. Verify the console output shows the full lifecycle across all layers.
