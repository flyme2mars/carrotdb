import socket
from typing import List, Optional

class Client:
    """
    The official Python client for CarrotDB.
    
    CarrotDB is a distributed, sharded Key-Value store. This client
    supports automatic namespacing, providing a multi-database experience
    on a single cluster.
    """

    def __init__(self, host: str = "localhost", port: int = 8000, database: str = "default"):
        """
        Initialize a new CarrotDB connection.

        Args:
            host (str): The hostname of the Carrot-Router. Defaults to "localhost".
            port (int): The port of the Carrot-Router. Defaults to 8000.
            database (str): The logical database (namespace) to use. Defaults to "default".
        """
        self.host = host
        self.port = port
        self.database = database
        self._sock: Optional[socket.socket] = None
        self._reader = None
        self._connect()

    def _connect(self):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.connect((self.host, self.port))
        self._reader = self._sock.makefile('r')

    def _prefix_key(self, key: str) -> str:
        return f"{self.database}:{key}"

    def _unprefix_key(self, prefixed_key: str) -> str:
        prefix = f"{self.database}:"
        if prefixed_key.startswith(prefix):
            return prefixed_key[len(prefix):]
        return prefixed_key

    def set(self, key: str, value: str) -> bool:
        """
        Store a value in the database.

        Args:
            key (str): The unique identifier for the data.
            value (str): The string value to store.

        Returns:
            bool: True if the operation was successful.

        Raises:
            Exception: If the server returns an error.
        """
        full_key = self._prefix_key(key)
        command = f"SET {full_key} {value}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            raise Exception(response)
        return True

    def get(self, key: str) -> Optional[str]:
        """
        Retrieve a value from the database.

        Args:
            key (str): The unique identifier to look up.

        Returns:
            Optional[str]: The stored value, or None if the key does not exist.
        """
        full_key = self._prefix_key(key)
        command = f"GET {full_key}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            if "key not found" in response:
                return None
            raise Exception(response)
        # Strip the '+' prefix
        return response[1:]

    def delete(self, key: str) -> bool:
        """
        Remove a key and its value from the database.

        Args:
            key (str): The unique identifier to remove.

        Returns:
            bool: True if the key was deleted (or didn't exist).
        """
        full_key = self._prefix_key(key)
        command = f"DELETE {full_key}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            raise Exception(response)
        return True

    def list_keys(self) -> List[str]:
        """
        List all keys present in the current database (namespace).

        Returns:
            List[str]: A list of key names (without the database prefix).
        """
        prefix = f"{self.database}:"
        command = f"KEYS {prefix}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            raise Exception(response)
        
        # Raw response is +key1 key2 key3
        raw_keys = response[1:].split()
        return [self._unprefix_key(k) for k in raw_keys]

    def close(self):
        """Close the connection to the database."""
        if self._sock:
            self._sock.close()
