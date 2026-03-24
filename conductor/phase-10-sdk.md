# CarrotDB Phase 10: Python SDK & Multi-Database Namespacing

The goal of this phase is to make CarrotDB usable from Python and implement a "Multi-Database" experience using **Logical Namespacing**.

## 🏗 Architecture

### 1. Logical Namespacing (Key Prefixing)
Instead of physical folders, we use key prefixes to isolate data.
- **Example:** If a user connects to the `billing` database, the SDK automatically turns `invoice:1` into `billing:invoice:1`.
- **Benefit:** Data remains perfectly distributed across shards without complex server-side changes.

### 2. Server-Side "KEYS" Support
To manage these databases, the server needs to be able to find all keys belonging to a namespace.
- **New Command:** `KEYS <prefix>`
- **Response:** A space-separated list of keys starting with that prefix.

### 3. Python SDK (`carrotdb-py`)
A clean, class-based library for Python developers.
- `db = Client(host="localhost", port=8000, database="my_app")`
- `db.set("key", "val")` -> sends `SET my_app:key val`
- `db.list_keys()` -> sends `KEYS my_app:` and strips the prefix for the user.

## 📂 Plan
1. **Update Engine:** Add a `KeysWithPrefix(prefix)` method to `internal/engine/engine.go`.
2. **Update Server:** Add the `KEYS` command to `internal/server/server.go`.
3. **Create Python Package:**
    - `sdk/python/carrotdb/client.py`: Core TCP logic and auto-prefixing.
    - `sdk/python/carrotdb/__init__.py`: Package entry point.
4. **Example Script:** Create `sdk/python/example.py` to demonstrate the multi-db usage.
5. **Verification:** Start a cluster, run the Python script, and verify that keys from different "databases" don't conflict.

## 🚀 Final Goal (v0.11.0)
A database that feels like a professional multi-tenant system to the end-user.
