package sharding

import (
	"testing"
)

func TestHashRing_AddGet(t *testing.T) {
	hr := NewHashRing(40)
	hr.AddShard("shard1")
	hr.AddShard("shard2")
	hr.AddShard("shard3")

	// Ensure different keys map to different shards
	keys := []string{"user:1", "user:2", "user:3", "user:100", "product:55"}
	shardsFound := make(map[string]bool)

	for _, key := range keys {
		shard := hr.GetShard(key)
		if shard == "" {
			t.Errorf("failed to find shard for key %s", key)
		}
		shardsFound[shard] = true
	}

	if len(shardsFound) < 2 {
		t.Errorf("expected keys to be distributed across multiple shards, but they all mapped to %v", shardsFound)
	}
}

func TestHashRing_Consistency(t *testing.T) {
	hr := NewHashRing(40)
	hr.AddShard("shard1")
	hr.AddShard("shard2")

	key := "my-secret-key"
	shard1 := hr.GetShard(key)
	shard2 := hr.GetShard(key)

	if shard1 != shard2 {
		t.Errorf("hashing is inconsistent: got %s then %s", shard1, shard2)
	}
}
