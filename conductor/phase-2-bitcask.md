# CarrotDB Phase 2: Bitcask Storage Engine

The goal of this phase is to move values from memory to disk while keeping only the keys and their locations in memory. This allows CarrotDB to store datasets larger than the available RAM.

## 🏗 Architecture
CarrotDB will adopt the **Bitcask** model.

### 1. In-Memory Index (KeyDir)
A hash map that stores the metadata for every key.
```go
type location struct {
    fileID    int
    offset    int64
    valueSize uint32
    timestamp int64
}
```
`keyDir map[string]location`

### 2. Disk Data Format (Binary)
Every write operation will be appended to the active log file as a binary record:
| Header (16 bytes) | Key (variable) | Value (variable) |
| :--- | :--- | :--- |
| CRC (4) \| Timestamp (4) \| KeySize (4) \| ValueSize (4) | Key Data | Value Data |

- **CRC (Cyclic Redundancy Check):** Ensures data wasn't corrupted during the write.
- **Timestamp:** For versioning/recovery.
- **KeySize/ValueSize:** Tells us how much to read from disk.

### 3. Core Operations
- **`Put(key, value)`:**
  1. Encode the record (CRC + TS + KeySize + ValSize + Key + Value).
  2. Append to the log file.
  3. Get the offset where the write started.
  4. Update `keyDir[key]` with the new `location`.
- **`Get(key)`:**
  1. Look up `location` in `keyDir`.
  2. If found, `Seek` to `offset + header_size + keysize` in the log file.
  3. `Read` exactly `valueSize` bytes.
- **`Delete(key)`:**
  1. Write a special "Tombstone" record (ValueSize = -1 or special flag).
  2. Remove the key from `keyDir`.

## 📂 Plan
1. **Define Binary Record:** Create a utility to encode/decode the header and record.
2. **Update `Engine` struct:** Change `map[string]string` to `map[string]location`.
3. **Refactor `Put`:** Implement binary appending and offset tracking.
4. **Refactor `Get`:** Implement random access reading using `Seek`.
5. **Update `restore`:** Update the startup recovery to parse binary records.
6. **Verification:** All Phase 1 tests must pass (interface remains unchanged).

## 🚀 Future (Phase 2.1)
- **Compaction:** Merge old log files to remove deleted/stale records.
- **Data File Rotation:** Switch to a new log file when the current one reaches a size limit (e.g., 512MB).
