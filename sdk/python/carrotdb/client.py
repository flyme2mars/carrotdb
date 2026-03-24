import socket

class Client:
    """CarrotDB Python Client with Multi-Database (Namespace) support."""

    def __init__(self, host="localhost", port=8000, database="default"):
        self.host = host
        self.port = port
        self.database = database
        self._sock = None
        self._connect()

    def _connect(self):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.connect((self.host, self.port))
        self._reader = self._sock.makefile('r')

    def _prefix_key(self, key):
        return f"{self.database}:{key}"

    def _unprefix_key(self, prefixed_key):
        prefix = f"{self.database}:"
        if prefixed_key.startswith(prefix):
            return prefixed_key[len(prefix):]
        return prefixed_key

    def set(self, key, value):
        full_key = self._prefix_key(key)
        command = f"SET {full_key} {value}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            raise Exception(response)
        return True

    def get(self, key):
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

    def delete(self, key):
        full_key = self._prefix_key(key)
        command = f"DELETE {full_key}\n"
        self._sock.sendall(command.encode())
        response = self._reader.readline().strip()
        if response.startswith("-ERROR"):
            raise Exception(response)
        return True

    def list_keys(self):
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
        if self._sock:
            self._sock.close()
