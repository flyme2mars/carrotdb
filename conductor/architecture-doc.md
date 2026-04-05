# Plan: Educational ARCHITECTURE.md for CarrotDB

This plan introduces a high-quality, educational `ARCHITECTURE.md` file that transforms CarrotDB into a readable textbook for distributed systems.

## Objective
Provide a conceptual and technical deep-dive into CarrotDB's design, explaining the trade-offs and algorithms used.

## Changes

### 1. `ARCHITECTURE.md` (New File)
- **Introduction:** Why CarrotDB? The "Unified Symmetric" design.
- **The Storage Engine (Bitcask):** 
    - Explaining the "Append-Only Log" vs "B-Trees."
    - Why sequential writes are faster than random writes.
    - How the in-memory `keyDir` works.
- **Distributed Consensus (Raft):**
    - The role of the **Leader** and the **Finite State Machine (FSM)**.
    - How **Strong Consistency** is achieved via the log replication quorum.
    - Snapshotting: Compacting history into a "picture."
- **Membership & Discovery (Gossip):**
    - Why we use **Memberlist (SWIM protocol)**.
    - **Eventual Consistency** for node health vs Strong Consistency for data.
- **Data Distribution (Consistent Hashing):**
    - The "Hash Ring" visual concept.
    - Why **Virtual Nodes** are critical for even load balancing.
    - How the Router maps a key to a shard.
- **The CAP Theorem:** Where CarrotDB sits on the triangle (CP vs AP).

### 2. `README.md`
- Link the new `ARCHITECTURE.md` as the "Recommended Reading" for students.

## Implementation Steps

### Phase 1: Storage & Single-Node Theory
1. Draft the Bitcask and Log-Structured storage sections.
2. Explain the CRC and Corruptions checks for educational value.

### Phase 2: Distributed Systems Theory
1. Draft the Raft and Consensus sections.
2. Draft the Gossip and Membership sections.
3. Compare and contrast the two protocols (Strong vs Eventual).

### Phase 3: Sharding & Routing Theory
1. Draft the Consistent Hashing section.
2. Explain the "Symmetric Architecture" (every node is a router).

### Phase 4: Final Integration
1. Review for clarity, educational tone, and technical accuracy.
2. Link from `README.md`.

## Verification
1. Review the file in a Markdown renderer.
2. Ensure all links to the source code (`internal/engine/engine.go`, etc.) are accurate.
