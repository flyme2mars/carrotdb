# CarrotDB Python SDK

The official Python client for [CarrotDB](https://github.com/flyme2mars/carrotdb).

## Installation

```bash
# Clone the repository
git clone https://github.com/flyme2mars/carrotdb.git
cd carrotdb/sdk/python

# (Optional) Install in editable mode
pip install -e .
```

## Quick Start

```python
from carrotdb import Client

# Connect to the Router (default port 8000)
# Provide a 'database' name to isolate your data from other apps
db = Client(host="localhost", port=8000, database="production_v1")

# Store a value
db.set("user:100", "Alice")

# Retrieve a value
name = db.get("user:100")
print(f"Hello, {name}!")

# List all keys in this database
keys = db.list_keys()
print(f"Total keys: {len(keys)}")

# Delete a key
db.delete("user:100")
```

## API Reference

### `Client(host="localhost", port=8000, database="default")`
Creates a connection to a CarrotDB cluster via the Router.

### `.set(key: str, value: str) -> bool`
Stores a string value. Returns `True` on success.

### `.get(key: str) -> Optional[str]`
Retrieves a value. Returns `None` if the key is not found.

### `.delete(key: str) -> bool`
Deletes a key.

### `.list_keys() -> List[str]`
Returns all keys belonging to the current `database` namespace.

### `.close()`
Closes the TCP connection.
