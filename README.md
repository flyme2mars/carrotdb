# 🥕 CarrotDB

A high-performance, scalable, and educational Key-Value database written in Go.

CarrotDB is designed to be simple to understand but powerful enough to handle massive datasets. It uses a **Log-Structured Storage Engine** (Bitcask-inspired) to ensure extreme write speeds and crash resilience.

## ✨ Features (v0.1.0)
- **Fast Writes:** Append-only logging ensures data is persisted to disk instantly.
- **Thread-Safe:** Built with Go's `sync.RWMutex` to handle concurrent operations.
- **Crash Recovery:** Automatically replays the write-ahead log (WAL) on startup to restore state.
- **Simple CLI:** Interactive command-line interface for direct data manipulation.

## 🚀 Quick Start

### 1. Installation

#### **The Portable Way (Recommended)**
Download the binary for your operating system from the [Releases](https://github.com/flyme2mars/carrotdb/releases) page.

#### **For Go Developers**
If you have Go installed, you can install CarrotDB directly:
```bash
go install github.com/flyme2mars/carrotdb/cmd/carrotdb@latest
```

### 2. Running CarrotDB
Unzip the downloaded binary and run:
```bash
./carrotdb
```

### 3. Usage
Once the CarrotDB prompt appears, you can use the following commands:
```bash
> SET user:1 Alice
OK
> GET user:1
Alice
> DELETE user:1
OK
> EXIT
```

## 🛠 System-Wide Installation (Optional)
To run `carrotdb` from any directory on your system:

**macOS / Linux:**
```bash
mv carrotdb /usr/local/bin/
```

**Windows:**
Add the folder containing `carrotdb.exe` to your system's **PATH** environment variable.

## 🗺 Roadmap
- [x] **Phase 1:** In-Memory Store + Append-Only Log (Current)
- [ ] **Phase 2:** Bitcask Storage Engine (Keys in RAM, Values on Disk)
- [ ] **Phase 3:** TCP Networking & Custom Protocol
- [ ] **Phase 4:** Distributed Consensus (Raft)

## 📄 License
MIT
