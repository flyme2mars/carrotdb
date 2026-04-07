# Plan: CarrotDB Lab Bench (Educational Benchmarking)

This plan introduces `carrotdb-bench`, a tool designed not just to measure performance, but to **teach the trade-offs** of distributed systems through quantitative experiments.

## Objective
Create a command-line tool that runs standardized workloads against a CarrotDB cluster and generates an educational report explaining the performance results based on the cluster's topology.

## 1. `cmd/carrotdb-bench/main.go` (New Tool)
- **Parameters:**
    - `-host`, `-port`: Target router address.
    - `-duration`: How long to run each experiment (default 10s).
    - `-concurrency`: Number of parallel workers (default 10).
- **Experiments:**
    - **Experiment 1: Write Throughput.** Measures how many `SET` operations the cluster can handle.
    - **Experiment 2: Key Discovery.** Measures the latency of the `KEYS` command.
- **The "Professor" Report:**
    - Calculates ops/sec and latency percentiles (P50, P99).
    - Queries the cluster state via `CLUSTER` command.
    - Generates a Markdown-style report with "The 'Why'" section for each result.

## 2. Protocol Enhancements
- Ensure the `CLUSTER` command provides enough metadata for the bench tool to understand the topology (shards, nodes per shard).

## Implementation Steps

### Phase 1: Core Benchmarking Engine
1. Implement the worker pool and timing logic.
2. Implement basic `SET` and `GET` measurement.

### Phase 2: Topology Awareness
1. Enhance the `CLUSTER` response parsing in the bench tool.
2. Map performance numbers to cluster size.

### Phase 3: Educational Reporting
1. Write the logic to generate the "Professor" commentary based on the results.
2. Add color-coded success/warning indicators.

## Verification
1. Run bench against a 1-node cluster.
2. Run bench against a 3-node cluster.
3. Verify that the report correctly identifies that more shards = higher `SET` throughput but potentially higher `KEYS` latency.
