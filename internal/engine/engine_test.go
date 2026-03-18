package engine

import (
	"os"
	"testing"
)

func TestEngine_PutGet(t *testing.T) {
	// Create a temporary file for the log
	tmpFile, err := os.CreateTemp("", "carrotdb-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	e, err := NewEngine(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	key := "user_1"
	value := "Alice"

	// Test Put
	err = e.Put(key, value)
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

func TestEngine_Recovery(t *testing.T) {
	// Create a temporary file for the log
	tmpFile, err := os.CreateTemp("", "carrotdb-recovery-*.log")
	if err != nil {
		t.Fatal(err)
	}
	logPath := tmpFile.Name()
	defer os.Remove(logPath)
	tmpFile.Close()

	// 1. Start engine and save data
	e1, err := NewEngine(logPath)
	if err != nil {
		t.Fatalf("failed to create engine 1: %v", err)
	}
	e1.Put("k1", "v1")
	e1.Put("k2", "v2")
	e1.Delete("k1") // k1 should be gone after recovery
	e1.Close()

	// 2. Start a NEW engine with the SAME file (Simulate Restart)
	e2, err := NewEngine(logPath)
	if err != nil {
		t.Fatalf("failed to create engine 2: %v", err)
	}
	defer e2.Close()

	// 3. Verify data was recovered
	val, err := e2.Get("k2")
	if err != nil || val != "v2" {
		t.Errorf("expected v2, got %s (err: %v)", val, err)
	}

	_, err = e2.Get("k1")
	if err == nil {
		t.Error("expected k1 to be deleted after recovery, but it exists")
	}
}

func TestEngine_GetNotFound(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "carrotdb-notfound-*.log")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	e, _ := NewEngine(tmpFile.Name())
	defer e.Close()

	_, err := e.Get("non_existent_key")
	if err == nil {
		t.Error("expected error for non-existent key, got nil")
	}
}
