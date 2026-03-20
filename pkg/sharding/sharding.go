package sharding

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashRing is a consistent hashing ring.
type HashRing struct {
	nodes           []uint32          // sorted hash values
	nodeMap         map[uint32]string // hash to shard ID
	replication     int               // number of virtual nodes per shard
}

// NewHashRing creates a new HashRing with the given replication factor.
func NewHashRing(replication int) *HashRing {
	return &HashRing{
		nodeMap:     make(map[uint32]string),
		replication: replication,
	}
}

// AddShard adds a shard (cluster) to the ring with its virtual nodes.
func (h *HashRing) AddShard(shardID string) {
	for i := 0; i < h.replication; i++ {
		hash := crc32.ChecksumIEEE([]byte(strconv.Itoa(i) + shardID))
		h.nodes = append(h.nodes, hash)
		h.nodeMap[hash] = shardID
	}
	sort.Slice(h.nodes, func(i, j int) bool {
		return h.nodes[i] < h.nodes[j]
	})
}

// GetShard finds the shard ID responsible for the given key.
func (h *HashRing) GetShard(key string) string {
	if len(h.nodes) == 0 {
		return ""
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	// Find the first virtual node with a hash >= key's hash
	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i] >= hash
	})

	// If we've reached the end of the ring, wrap around to the first node
	if idx == len(h.nodes) {
		idx = 0
	}

	return h.nodeMap[h.nodes[idx]]
}
