package main

import (
	"carrotdb/internal/engine"
	"carrotdb/internal/server"
	"log"
	"os"
)

func main() {
	// Ensure the data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize the storage engine
	db, err := engine.NewEngine("data/carrotdb.log")
	if err != nil {
		log.Fatalf("failed to start engine: %v", err)
	}
	defer db.Close()

	// Initialize and start the TCP server
	s := server.NewServer(":6379", db)
	if err := s.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
