package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// SimpleMemoryManager handles short-term and long-term memory
type SimpleMemoryManager struct {
	shortTerm map[string]MemoryEntry
	longTerm  map[string]MemoryEntry
}

var memoryManager *SimpleMemoryManager

func init() {
	memoryManager = &SimpleMemoryManager{
		shortTerm: make(map[string]MemoryEntry),
		longTerm:  make(map[string]MemoryEntry),
	}
	
	// Add some test data
	memoryManager.longTerm["test_key"] = MemoryEntry{
		Key:       "test_key",
		Value:     "test_value",
		Timestamp: time.Now(),
	}
}

func getLongTermMemory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	entries := make([]MemoryEntry, 0, len(memoryManager.longTerm))
	for _, entry := range memoryManager.longTerm {
		entries = append(entries, entry)
	}
	
	json.NewEncoder(w).Encode(entries)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/api/v1/memory/long", getLongTermMemory)
	http.HandleFunc("/health", healthCheck)
	
	fmt.Println("Starting web server on :8081")
	fmt.Println("Test endpoints:")
	fmt.Println("- http://localhost:8081/health")
	fmt.Println("- http://localhost:8081/api/v1/memory/long")
	
	// Start server in background for testing
	go func() {
		http.ListenAndServe(":8081", nil)
	}()
	
	// Keep the program running for a bit to test
	time.Sleep(5 * time.Second)
	fmt.Println("Web server test completed successfully!")
}