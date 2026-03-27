# Plan: Production-Grade Docker Integration for CarrotDB

This plan outlines the "right way" to containerize CarrotDB, ensuring that it remains performant, persistent, and easy to orchestrate in both local and distributed environments.

## 1. Multi-Stage Dockerfile (Efficiency & Security)
We will use a multi-stage build to keep the final image minimal and secure.

### Stage 1: Build
-   Use `golang:1.24-alpine` as the builder.
-   Compile both `carrotdb-server` and the `carrotdb` CLI.
-   Enable `CGO_ENABLED=0` for a truly static binary that runs anywhere.

### Stage 2: Final Image
-   Use `alpine:latest` or `scratch` (if no shell is needed) for the smallest attack surface.
-   Include `ca-certificates` for secure communication.
-   Expose ports: `6379` (API), `7000` (Raft), `8000` (Router), `8080` (Dashboard), `9000` (Gossip).
-   Set a non-root user for security.

## 2. Dynamic Configuration via Environment Variables
Hardcoding flags in `ENTRYPOINT` is brittle. We will modify `main.go` (or use a wrapper script) to prioritize Environment Variables:
-   `CARROT_ID`, `CARROT_SHARD`, `CARROT_JOIN`, `CARROT_GOSSIP_SEED`, etc.

## 3. Persistent Data Strategy
Docker containers are ephemeral. CarrotDB's Raft logs and Bitcask data must survive restarts.
-   **Volume Mapping:** The container will expect a volume at `/data`.
-   **Ownership:** Ensure the non-root user in the container has write permissions to the mounted volume.

## 4. Orchestration Scenarios

### Scenario A: Local Development (Docker Compose)
A `docker-compose.yaml` that spins up a 3-node cluster with one command.
-   Node 1: Seed node.
-   Nodes 2 & 3: Join the cluster automatically using the service name `node1` as the gossip seed.

### Scenario B: Production (Kubernetes / Cloud)
-   Use **StatefulSets** to ensure each node keeps its identity (e.g., `carrotdb-0`, `carrotdb-1`).
-   Use **Persistent Volume Claims (PVCs)** for stable storage.
-   Headless services for stable network identities (DNS) required by Raft.

## 5. Implementation Steps

### Phase 1: The Dockerfile
1.  Create `Dockerfile` at the project root.
2.  Implement multi-stage build for `carrotdb` and `carrotdb-server`.

### Phase 2: Configuration Overhaul
1.  Update `cmd/carrotdb-server/main.go` to check for environment variables if flags are missing.
2.  Add a default health check endpoint (or use the existing Dashboard API).

### Phase 3: The "One-Click" Cluster
1.  Create `docker-compose.yaml`.
2.  Define a standard network and volume strategy.

## 6. Verification & Testing
1.  **Build Test:** `docker build -t carrotdb:latest .`
2.  **Cluster Test:** `docker-compose up -d` and verify shard health via the dashboard.
3.  **Persistence Test:** `docker-compose stop`, then `start`, and verify data is still there.
4.  **CLI Test:** Use the containerized CLI to connect to the cluster: `docker exec -it node1 carrotdb SET key val`.

## 7. Best Approach Recommendation
**Recommendation:** Use the **Stateful Symmetric** approach. Since every CarrotDB node is a router, we don't need a separate "Load Balancer" container. A simple Round-Robin DNS or Service IP can point to any node in the cluster, and it will correctly route the request to the leader.
