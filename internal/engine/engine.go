package engine

import (
	"errors"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

// Engine is the core storage engine.
type Engine struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewEngine creates a new instance of the Engine.
func NewEngine() *Engine {
	return &Engine{
		data: make(map[string]string),
	}
}

// Put adds a new key-value pair or updates an existing one.
func (e *Engine) Put(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = value
	return nil
}

// Get retrieves a value associated with the given key.
func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	val, ok := e.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	return val, nil
}

// Delete removes a key and its associated value.
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.data, key)
	return nil
}
