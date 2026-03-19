package engine

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"
	"time"
)

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrDataCorrupted   = errors.New("data corrupted")
	ErrInvalidRecord   = errors.New("invalid record")
)

const (
	headerSize = 16 // 4 (CRC) + 4 (TS) + 4 (KS) + 4 (VS)
	tombstone  = ^uint32(0) // special value size for deletions
)

// location stores the offset and size of a value on disk.
type location struct {
	offset    int64
	valueSize uint32
	timestamp uint32
}

// Engine is the core storage engine using the Bitcask model.
type Engine struct {
	mu     sync.RWMutex
	keyDir map[string]location
	log    *os.File
}

// NewEngine creates a new instance of the Engine and restores its state from the log.
func NewEngine(logPath string) (*Engine, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	e := &Engine{
		keyDir: make(map[string]location),
		log:    file,
	}

	if err := e.restore(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to restore engine state: %w", err)
	}

	return e, nil
}

// restore reads the log file from the beginning and re-builds the in-memory keyDir.
func (e *Engine) restore() error {
	if _, err := e.log.Seek(0, 0); err != nil {
		return err
	}

	for {
		offset, err := e.log.Seek(0, 1) // current offset
		if err != nil {
			return err
		}

		header := make([]byte, headerSize)
		_, err = io.ReadFull(e.log, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		crc := binary.LittleEndian.Uint32(header[0:4])
		timestamp := binary.LittleEndian.Uint32(header[4:8])
		keySize := binary.LittleEndian.Uint32(header[8:12])
		valueSize := binary.LittleEndian.Uint32(header[12:16])

		key := make([]byte, keySize)
		if _, err := io.ReadFull(e.log, key); err != nil {
			return err
		}

		var value []byte
		if valueSize != tombstone {
			value = make([]byte, valueSize)
			if _, err := io.ReadFull(e.log, value); err != nil {
				return err
			}
		}

		// Verify CRC
		calculatedCRC := crc32.ChecksumIEEE(header[4:])
		calculatedCRC = crc32.Update(calculatedCRC, crc32.IEEETable, key)
		if valueSize != tombstone {
			calculatedCRC = crc32.Update(calculatedCRC, crc32.IEEETable, value)
		}

		if crc != calculatedCRC {
			return ErrDataCorrupted
		}

		if valueSize == tombstone {
			delete(e.keyDir, string(key))
		} else {
			e.keyDir[string(key)] = location{
				offset:    offset,
				valueSize: valueSize,
				timestamp: timestamp,
			}
		}
	}

	// Seek back to the end for future appends
	_, err := e.log.Seek(0, 2)
	return err
}

// Put adds a new key-value pair, writes to the log, and updates the index.
func (e *Engine) Put(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	timestamp := uint32(time.Now().Unix())
	keySize := uint32(len(key))
	valueSize := uint32(len(value))

	// Get current offset for the keyDir
	offset, err := e.log.Seek(0, 1)
	if err != nil {
		return err
	}

	// Prepare record
	header := make([]byte, headerSize)
	binary.LittleEndian.PutUint32(header[4:8], timestamp)
	binary.LittleEndian.PutUint32(header[8:12], keySize)
	binary.LittleEndian.PutUint32(header[12:16], valueSize)

	crc := crc32.ChecksumIEEE(header[4:])
	crc = crc32.Update(crc, crc32.IEEETable, []byte(key))
	crc = crc32.Update(crc, crc32.IEEETable, []byte(value))
	binary.LittleEndian.PutUint32(header[0:4], crc)

	// Write record: header + key + value
	if _, err := e.log.Write(header); err != nil {
		return err
	}
	if _, err := e.log.WriteString(key); err != nil {
		return err
	}
	if _, err := e.log.WriteString(value); err != nil {
		return err
	}

	if err := e.log.Sync(); err != nil {
		return err
	}

	// Update index
	e.keyDir[key] = location{
		offset:    offset,
		valueSize: valueSize,
		timestamp: timestamp,
	}

	return nil
}

// Get retrieves a value associated with the given key by reading from disk.
func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	loc, ok := e.keyDir[key]
	e.mu.RUnlock()

	if !ok {
		return "", ErrKeyNotFound
	}

	// Seek to the value part of the record
	// Record is: CRC(4) | TS(4) | KS(4) | VS(4) | Key | Value
	// valueOffset = loc.offset + headerSize + len(key)
	valueOffset := loc.offset + int64(headerSize) + int64(len(key))

	valBytes := make([]byte, loc.valueSize)
	_, err := e.log.ReadAt(valBytes, valueOffset)
	if err != nil {
		return "", err
	}

	return string(valBytes), nil
}

// Delete removes a key, writes a tombstone to the log, and updates the index.
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.keyDir[key]; !ok {
		return nil // key doesn't exist, nothing to do
	}

	timestamp := uint32(time.Now().Unix())
	keySize := uint32(len(key))
	valueSize := tombstone

	// Prepare tombstone record
	header := make([]byte, headerSize)
	binary.LittleEndian.PutUint32(header[4:8], timestamp)
	binary.LittleEndian.PutUint32(header[8:12], keySize)
	binary.LittleEndian.PutUint32(header[12:16], valueSize)

	crc := crc32.ChecksumIEEE(header[4:])
	crc = crc32.Update(crc, crc32.IEEETable, []byte(key))
	binary.LittleEndian.PutUint32(header[0:4], crc)

	if _, err := e.log.Write(header); err != nil {
		return err
	}
	if _, err := e.log.WriteString(key); err != nil {
		return err
	}

	if err := e.log.Sync(); err != nil {
		return err
	}

	delete(e.keyDir, key)
	return nil
}

// Compact merges the log file to remove stale and deleted keys.
func (e *Engine) Compact() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Create a temporary file for merging
	mergePath := e.log.Name() + ".merge"
	mergeFile, err := os.OpenFile(mergePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create merge file: %w", err)
	}

	newKeyDir := make(map[string]location)
	var newOffset int64

	// 2. Iterate over the current valid keys in keyDir
	for key, loc := range e.keyDir {
		// Read value from current log
		valueOffset := loc.offset + int64(headerSize) + int64(len(key))
		valBytes := make([]byte, loc.valueSize)
		if _, err := e.log.ReadAt(valBytes, valueOffset); err != nil {
			mergeFile.Close()
			os.Remove(mergePath)
			return fmt.Errorf("failed to read value for key %s during compaction: %w", key, err)
		}

		// Write to merge file
		timestamp := loc.timestamp
		keySize := uint32(len(key))
		valueSize := loc.valueSize

		header := make([]byte, headerSize)
		binary.LittleEndian.PutUint32(header[4:8], timestamp)
		binary.LittleEndian.PutUint32(header[8:12], keySize)
		binary.LittleEndian.PutUint32(header[12:16], valueSize)

		crc := crc32.ChecksumIEEE(header[4:])
		crc = crc32.Update(crc, crc32.IEEETable, []byte(key))
		crc = crc32.Update(crc, crc32.IEEETable, valBytes)
		binary.LittleEndian.PutUint32(header[0:4], crc)

		if _, err := mergeFile.Write(header); err != nil {
			mergeFile.Close()
			os.Remove(mergePath)
			return err
		}
		if _, err := mergeFile.WriteString(key); err != nil {
			mergeFile.Close()
			os.Remove(mergePath)
			return err
		}
		if _, err := mergeFile.Write(valBytes); err != nil {
			mergeFile.Close()
			os.Remove(mergePath)
			return err
		}

		// Update the new keyDir with the offset in the merge file
		newKeyDir[key] = location{
			offset:    newOffset,
			valueSize: valueSize,
			timestamp: timestamp,
		}

		newOffset += int64(headerSize) + int64(keySize) + int64(valueSize)
	}

	if err := mergeFile.Sync(); err != nil {
		mergeFile.Close()
		os.Remove(mergePath)
		return err
	}

	// 3. Swap files
	oldPath := e.log.Name()
	e.log.Close()
	mergeFile.Close()

	if err := os.Rename(mergePath, oldPath); err != nil {
		return fmt.Errorf("failed to replace old log with merge file: %w", err)
	}

	// Re-open the newly compacted log
	newLog, err := os.OpenFile(oldPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to re-open compacted log: %w", err)
	}

	e.log = newLog
	e.keyDir = newKeyDir

	return nil
}

// Close closes the engine's log file safely.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.log.Close()
}
