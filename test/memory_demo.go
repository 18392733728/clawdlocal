package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       *time.Duration `json:"ttl,omitempty"`
}

// SimpleMemoryManager handles short-term and long-term memory
type SimpleMemoryManager struct {
	shortTerm map[string]MemoryEntry
	longTerm  map[string]MemoryEntry
	mutex     sync.RWMutex
}

// NewSimpleMemoryManager creates a new memory manager
func NewSimpleMemoryManager() *SimpleMemoryManager {
	return &SimpleMemoryManager{
		shortTerm: make(map[string]MemoryEntry),
		longTerm:  make(map[string]MemoryEntry),
	}
}

// SetShortTermMemory stores a value in short-term memory with optional TTL
func (mm *SimpleMemoryManager) SetShortTermMemory(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	entry := MemoryEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	if ttl > 0 {
		entry.TTL = &ttl
	}

	mm.shortTerm[key] = entry
	return nil
}

// GetShortTermMemory retrieves a value from short-term memory
func (mm *SimpleMemoryManager) GetShortTermMemory(ctx context.Context, key string) (interface{}, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entry, exists := mm.shortTerm[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Check TTL
	if entry.TTL != nil {
		if time.Since(entry.Timestamp) > *entry.TTL {
			delete(mm.shortTerm, key)
			return nil, fmt.Errorf("key expired: %s", key)
		}
	}

	return entry.Value, nil
}

// SetLongTermMemory stores a value in long-term memory
func (mm *SimpleMemoryManager) SetLongTermMemory(ctx context.Context, key string, value interface{}) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	entry := MemoryEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	mm.longTerm[key] = entry
	return nil
}

// GetLongTermMemory retrieves a value from long-term memory
func (mm *SimpleMemoryManager) GetLongTermMemory(ctx context.Context, key string) (interface{}, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entry, exists := mm.longTerm[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return entry.Value, nil
}

// GetAllShortTerm returns all short-term memory entries
func (mm *SimpleMemoryManager) GetAllShortTerm(ctx context.Context) ([]MemoryEntry, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entries := make([]MemoryEntry, 0, len(mm.shortTerm))
	for _, entry := range mm.shortTerm {
		// Check TTL
		if entry.TTL != nil {
			if time.Since(entry.Timestamp) > *entry.TTL {
				continue // Skip expired entries
			}
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAllLongTerm returns all long-term memory entries
func (mm *SimpleMemoryManager) GetAllLongTerm(ctx context.Context) ([]MemoryEntry, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entries := make([]MemoryEntry, 0, len(mm.longTerm))
	for _, entry := range mm.longTerm {
		entries = append(entries, entry)
	}

	return entries, nil
}

// SaveToFile saves memory to a file
func (mm *SimpleMemoryManager) SaveToFile(filename string) error {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	data := map[string]interface{}{
		"short_term": mm.shortTerm,
		"long_term":  mm.longTerm,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// LoadFromFile loads memory from a file
func (mm *SimpleMemoryManager) LoadFromFile(filename string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, that's fine
		}
		return err
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// Convert back to MemoryEntry maps
	if shortTerm, ok := data["short_term"].(map[string]interface{}); ok {
		mm.shortTerm = make(map[string]MemoryEntry)
		for key, value := range shortTerm {
			if entryMap, ok := value.(map[string]interface{}); ok {
				entry := MemoryEntry{
					Key: key,
				}
				if val, ok := entryMap["Value"]; ok {
					entry.Value = val
				}
				if ts, ok := entryMap["Timestamp"].(string); ok {
					if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
						entry.Timestamp = parsed
					}
				}
				if ttl, ok := entryMap["TTL"].(float64); ok {
					duration := time.Duration(ttl) * time.Second
					entry.TTL = &duration
				}
				mm.shortTerm[key] = entry
			}
		}
	}

	if longTerm, ok := data["long_term"].(map[string]interface{}); ok {
		mm.longTerm = make(map[string]MemoryEntry)
		for key, value := range longTerm {
			if entryMap, ok := value.(map[string]interface{}); ok {
				entry := MemoryEntry{
					Key: key,
				}
				if val, ok := entryMap["Value"]; ok {
					entry.Value = val
				}
				if ts, ok := entryMap["Timestamp"].(string); ok {
					if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
						entry.Timestamp = parsed
					}
				}
				mm.longTerm[key] = entry
			}
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	mm := NewSimpleMemoryManager()

	// Test short-term memory with TTL
	fmt.Println("Testing short-term memory...")
	err := mm.SetShortTermMemory(ctx, "test_key", "test_value", 10*time.Second)
	if err != nil {
		fmt.Printf("Error setting short-term memory: %v\n", err)
		return
	}

	value, err := mm.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		fmt.Printf("Error getting short-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved short-term memory: %v\n", value)

	// Test long-term memory
	fmt.Println("Testing long-term memory...")
	err = mm.SetLongTermMemory(ctx, "long_key", "long_value")
	if err != nil {
		fmt.Printf("Error setting long-term memory: %v\n", err)
		return
	}

	value, err = mm.GetLongTermMemory(ctx, "long_key")
	if err != nil {
		fmt.Printf("Error getting long-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved long-term memory: %v\n", value)

	// Test saving and loading
	fmt.Println("Testing save/load...")
	err = mm.SaveToFile("test_memory.json")
	if err != nil {
		fmt.Printf("Error saving memory: %v\n", err)
		return
	}

	newMM := NewSimpleMemoryManager()
	err = newMM.LoadFromFile("test_memory.json")
	if err != nil {
		fmt.Printf("Error loading memory: %v\n", err)
		return
	}

	value, err = newMM.GetLongTermMemory(ctx, "long_key")
	if err != nil {
		fmt.Printf("Error getting loaded long-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved loaded long-term memory: %v\n", value)

	fmt.Println("All tests passed!")
}