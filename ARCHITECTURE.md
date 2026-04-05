# CarrotDB Architecture: A Masterclass in Distributed Systems

CarrotDB is more than just a database; it is a **living textbook** designed to teach the fundamental trade-offs of modern distributed storage. This document explains the "Why" and "How" behind every architectural decision.

---

## 1. Storage: The Bitcask Model
Traditional databases (like PostgreSQL or MySQL) use **B-Trees**. B-Trees are complex and require "random" disk access, which can be slow. CarrotDB uses a **Log-Structured Merge-ish (Bitcask)** model.

### 📝 The Append-Only Log
Every write (`SET`) in CarrotDB is simply appended to the end of a file.
*   **Why?** Sequential writes are significantly faster than random writes because the disk head doesn't have to "seek" all over the platter.
*   **The Trade-off:** The file grows forever. To solve this, CarrotDB implements **Compaction** (merging old logs and removing stale keys).

### 🧠 The KeyDir (In-Memory Index)
To make reads fast, CarrotDB keeps all **Keys** in memory, but all **Values** on disk.
*   **Storage:** `map[string]Location{offset, size}`.
*   **The Result:** A `GET` request is always exactly **one disk seek**. This provides O(1) performance while allowing you to store datasets larger than your available RAM.

---

## 2. Consistency: The Raft Consensus Algorithm
In a distributed system, nodes can fail or network messages can be lost. How do we ensure everyone agrees on the data? CarrotDB uses **Raft**.

### 👑 Leaders and Followers
A CarrotDB shard elects a single **Leader**. All writes go to the Leader.
*   **The Log Replication:** The Leader proposes a change to its **Followers**. Once a majority (quorum) acknowledges the change, it is "committed."
*   **Strong Consistency:** Raft ensures that as long as a majority of nodes are alive, the database remains correct and linearizable.

### 📸 Snapshotting & FSM
The **Finite State Machine (FSM)** is the "brain" that turns Raft logs into actual data.
*   **Snapshotting:** Instead of replaying millions of logs on restart, CarrotDB takes a "picture" of the state and clears the log history. This is essential for long-term durability.

---

## 3. Membership: The Gossip Protocol (SWIM)
How does Node A know that Node B just joined the cluster without a central "master" server? CarrotDB uses **Gossip**.

*   **SWIM Protocol:** Nodes periodically "ping" a few random neighbors. If a neighbor doesn't respond, the news of its "suspected" failure spreads through the cluster like a rumor (gossip).
*   **Discovery:** When you join a cluster, you only need to know **one** existing node's address. The Gossip protocol will automatically introduce you to everyone else.

---

## 4. Scaling: Consistent Hashing
CarrotDB scales horizontally by splitting data into **Shards**. To decide which shard owns a key, we use **Consistent Hashing**.

### 💍 The Hash Ring
Imagine a circle (a ring) where every point is a number from 0 to 2^32.
1.  We hash every **Shard ID** to a point on the ring.
2.  We hash every **Key** to a point on the ring.
3.  The key is owned by the first Shard it finds while moving clockwise around the ring.

### 🌀 Virtual Nodes
To prevent "hotspots" (one shard getting too much data), CarrotDB uses **Virtual Nodes**. Every physical shard is represented multiple times on the ring, ensuring a statistically even distribution of data.

---

## 5. Symmetry: The Unified Node Architecture
Unlike other databases that have separate "Router" and "Storage" binaries, every CarrotDB node is **Symmetric**.
*   **Every node is a Router:** Any node can receive a request from a client.
*   **Every node is a Storage Engine:** Every node participates in a Raft group.
*   **The Result:** There is no "single point of failure." You can point your application at any node in the cluster, and it will correctly route your request to the leader of the appropriate shard.

---

## ⚖️ The CAP Theorem
In distributed systems, you can only pick two: **C**onsistency, **A**vailability, or **P**artition Tolerance.
*   **CarrotDB is a CP System:** We prioritize **Consistency** and **Partition Tolerance**. During a network split, CarrotDB will stop accepting writes on the minority side to ensure that your data never becomes corrupted or out-of-sync.

---
*Ready to dive deeper? Check the code in `internal/engine/`, `internal/server/`, and `pkg/sharding/` to see these concepts in action.*
