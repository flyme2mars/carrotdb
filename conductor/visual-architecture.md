# Plan: Visual & Conceptual Mastery (ARCHITECTURE.md Upgrade)

This plan transforms `ARCHITECTURE.md` into a rich, visual guide using Mermaid diagrams and deep-dive conceptual sections to maximize its educational value.

## Objective
Make `ARCHITECTURE.md` the definitive visual resource for learning distributed systems by adding diagrams for every major concept (Bitcask, Raft, Gossip, Hashing).

## Changes

### 1. `ARCHITECTURE.md` (Visual Overhaul)
- **Visual Bitcask:** Add a diagram showing the relationship between the **Append-Only Log** (on disk) and the **KeyDir** (in memory).
- **Visual Raft:** Add a diagram illustrating the **Log Replication Quorum** (Leader proposing to Followers).
- **Visual Hashing:** Add a **Hash Ring diagram** showing how Keys and Shards (with Virtual Nodes) are mapped.
- **Visual Symmetry:** Diagram showing how any Node acts as a **Gateway/Router** to the rest of the cluster.
- **Deep-Dive Sections:** Add "Pro-Tips" and "Common Pitfalls" boxes for each section to provide advanced insights.

## Implementation Steps

### Phase 1: Storage Visualization
1. Add a Mermaid diagram for the Bitcask "Write Path" (Memory -> Disk).
2. Add a Mermaid diagram for the Bitcask "Read Path" (Memory Offset -> Disk Read).

### Phase 2: Consensus Visualization
1. Add a Mermaid diagram for the Raft "Consensus Cycle" (Client -> Leader -> Followers -> Commit).
2. Explain the "Split-Brain" scenario visually.

### Phase 3: Sharding Visualization
1. Add a Mermaid diagram for the "Consistent Hash Ring."
2. Illustrate how **Virtual Nodes** prevent data skew.

### Phase 4: Network & Discovery Visualization
1. Add a Mermaid diagram for the **Gossip (SWIM) protocol** ping-pong.
2. Illustrate the **Symmetric Request Flow** (Client -> Random Node -> Correct Leader).

## Verification
1. Review the Mermaid diagrams in a renderer to ensure they are clean and accurate.
2. Verify that the diagrams align perfectly with the source code implementations.
