package engine

import (
	"testing"
)

func TestEngine_PutGet(t *testing.T) {
	e := NewEngine()

	key := "user_1"
	value := "Alice"

	// Test Put
	err := e.Put(key, value)
	if err != nil {
		t.Fatalf("failed to put key: %v", err)
	}

	// Test Get
	got, err := e.Get(key)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}

	if got != value {
		t.Errorf("expected %s, got %s", value, got)
	}
}

func TestEngine_GetNotFound(t *testing.T) {
	e := NewEngine()

	_, err := e.Get("non_existent_key")
	if err == nil {
		t.Error("expected error for non-existent key, got nil")
	}
}
