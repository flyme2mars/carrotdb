package engine

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

const (
	opSet = "SET"
	opDel = "DEL"
)

// Engine is the core storage engine with durability support.
type Engine struct {
	mu   sync.RWMutex
	data map[string]string
	log  *os.File
}

// NewEngine creates a new instance of the Engine and restores its state from the log.
func NewEngine(logPath string) (*Engine, error) {
	// Open or create the log file for appending
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	e := &Engine{
		data: make(map[string]string),
		log:  file,
	}

	// Restore state by replaying the log
	if err := e.restore(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to restore engine state: %w", err)
	}

	return e, nil
}

// restore reads the log file from the beginning and re-builds the in-memory map.
func (e *Engine) restore() error {
	// Seek to the beginning of the file to read it
	if _, err := e.log.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(e.log)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		op := parts[0]
		key := parts[1]

		switch op {
		case opSet:
			if len(parts) == 3 {
				e.data[key] = parts[2]
			}
		case opDel:
			delete(e.data, key)
		}
	}

	// Seek back to the end of the file so new writes are appended
	_, err := e.log.Seek(0, 2)
	return err
}

// Put adds a new key-value pair, writes to the log, and updates memory.
func (e *Engine) Put(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Write to Log (Durability)
	entry := fmt.Sprintf("%s|%s|%s\n", opSet, key, value)
	if _, err := e.log.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to log: %w", err)
	}
	// Force flush to disk
	if err := e.log.Sync(); err != nil {
		return fmt.Errorf("failed to sync log: %w", err)
	}

	// 2. Update Memory
	e.data[key] = value
	return nil
}

// Get retrieves a value associated with the given key from memory.
func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	val, ok := e.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	return val, nil
}

// Delete removes a key, writes the deletion to the log, and updates memory.
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Write to Log (Durability)
	entry := fmt.Sprintf("%s|%s\n", opDel, key)
	if _, err := e.log.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to log: %w", err)
	}
	if err := e.log.Sync(); err != nil {
		return fmt.Errorf("failed to sync log: %w", err)
	}

	// 2. Update Memory
	delete(e.data, key)
	return nil
}

// Close closes the engine's log file safely.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.log.Close()
}
