package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       *time.Duration `json:"ttl,omitempty"` // Time to live (optional)
}

// MemoryManager manages both short-term and long-term memory
type MemoryManager struct {
	shortTerm map[string]*MemoryEntry
	longTerm  map[string]*MemoryEntry
	mutex     sync.RWMutex
	logger    *logrus.Logger
	config    *MemoryConfig
}

// MemoryConfig holds memory configuration
type MemoryConfig struct {
	ShortTermCapacity int           `yaml:"short_term_capacity"`
	LongTermFile      string        `yaml:"long_term_file"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(logger *logrus.Logger, config *MemoryConfig) (*MemoryManager, error) {
	if config == nil {
		config = &MemoryConfig{
			ShortTermCapacity: 1000,
			LongTermFile:      "memory/long_term.json",
			CleanupInterval:   5 * time.Minute,
		}
	}

	mm := &MemoryManager{
		shortTerm: make(map[string]*MemoryEntry),
		longTerm:  make(map[string]*MemoryEntry),
		logger:    logger,
		config:    config,
	}

	// Load long-term memory from file
	if err := mm.loadLongTermMemory(); err != nil {
		logger.WithError(err).Warn("Failed to load long-term memory")
	}

	return mm, nil
}

// SetShortTermMemory stores a value in short-term memory
func (mm *MemoryManager) SetShortTermMemory(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	entry := &MemoryEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	if ttl > 0 {
		entry.TTL = &ttl
	}

	mm.shortTerm[key] = entry

	// Enforce capacity limit
	if len(mm.shortTerm) > mm.config.ShortTermCapacity {
		mm.evictOldestShortTerm()
	}

	return nil
}

// GetShortTermMemory retrieves a value from short-term memory
func (mm *MemoryManager) GetShortTermMemory(ctx context.Context, key string) (interface{}, bool, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entry, exists := mm.shortTerm[key]
	if !exists {
		return nil, false, nil
	}

	// Check if entry has expired
	if entry.TTL != nil {
		if time.Since(entry.Timestamp) > *entry.TTL {
			// Entry has expired, remove it
			mm.mutex.RUnlock()
			mm.mutex.Lock()
			delete(mm.shortTerm, key)
			mm.mutex.Unlock()
			mm.mutex.RLock()
			return nil, false, nil
		}
	}

	return entry.Value, true, nil
}

// SetLongTermMemory stores a value in long-term memory
func (mm *MemoryManager) SetLongTermMemory(ctx context.Context, key string, value interface{}) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	entry := &MemoryEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	mm.longTerm[key] = entry

	// Save to file immediately
	return mm.saveLongTermMemory()
}

// GetLongTermMemory retrieves a value from long-term memory
func (mm *MemoryManager) GetLongTermMemory(ctx context.Context, key string) (interface{}, bool, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entry, exists := mm.longTerm[key]
	if !exists {
		return nil, false, nil
	}

	return entry.Value, true, nil
}

// GetAllShortTermMemory returns all short-term memory entries
func (mm *MemoryManager) GetAllShortTermMemory(ctx context.Context) ([]*MemoryEntry, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entries := make([]*MemoryEntry, 0, len(mm.shortTerm))
	for _, entry := range mm.shortTerm {
		// Skip expired entries
		if entry.TTL != nil && time.Since(entry.Timestamp) > *entry.TTL {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAllLongTermMemory returns all long-term memory entries
func (mm *MemoryManager) GetAllLongTermMemory(ctx context.Context) ([]*MemoryEntry, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	entries := make([]*MemoryEntry, 0, len(mm.longTerm))
	for _, entry := range mm.longTerm {
		entries = append(entries, entry)
	}

	return entries, nil
}

// evictOldestShortTerm removes the oldest entry from short-term memory
func (mm *MemoryManager) evictOldestShortTerm() {
	if len(mm.shortTerm) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range mm.shortTerm {
		if entry.Timestamp.Before(oldestTime) {
			oldestTime = entry.Timestamp
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(mm.shortTerm, oldestKey)
		mm.logger.Debugf("Evicted oldest short-term memory entry: %s", oldestKey)
	}
}

// saveLongTermMemory saves long-term memory to file
func (mm *MemoryManager) saveLongTermMemory() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(mm.config.LongTermFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(mm.longTerm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal long-term memory: %w", err)
	}

	if err := os.WriteFile(mm.config.LongTermFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write long-term memory file: %w", err)
	}

	mm.logger.Debugf("Saved long-term memory to %s", mm.config.LongTermFile)
	return nil
}

// loadLongTermMemory loads long-term memory from file
func (mm *MemoryManager) loadLongTermMemory() error {
	data, err := os.ReadFile(mm.config.LongTermFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, that's okay
			return nil
		}
		return fmt.Errorf("failed to read long-term memory file: %w", err)
	}

	if len(data) == 0 {
		// Empty file, that's okay
		return nil
	}

	var longTerm map[string]*MemoryEntry
	if err := json.Unmarshal(data, &longTerm); err != nil {
		return fmt.Errorf("failed to unmarshal long-term memory: %w", err)
	}

	mm.longTerm = longTerm
	mm.logger.Infof("Loaded %d long-term memory entries from %s", len(longTerm), mm.config.LongTermFile)
	return nil
}

// StartCleanup starts the background cleanup process
func (mm *MemoryManager) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(mm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.cleanupExpiredEntries()
		case <-ctx.Done():
			return
		}
	}
}

// cleanupExpiredEntries removes expired entries from short-term memory
func (mm *MemoryManager) cleanupExpiredEntries() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	now := time.Now()
	removed := 0

	for key, entry := range mm.shortTerm {
		if entry.TTL != nil && now.Sub(entry.Timestamp) > *entry.TTL {
			delete(mm.shortTerm, key)
			removed++
		}
	}

	if removed > 0 {
		mm.logger.Debugf("Cleaned up %d expired short-term memory entries", removed)
	}
}

// SearchMemory searches for entries containing the query string
func (mm *MemoryManager) SearchMemory(ctx context.Context, query string, searchLongTerm bool) ([]*MemoryEntry, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var results []*MemoryEntry
	query = strings.ToLower(query)

	// Search short-term memory
	for _, entry := range mm.shortTerm {
		if entry.TTL != nil && time.Since(entry.Timestamp) > *entry.TTL {
			continue // Skip expired entries
		}
		if mm.entryMatchesQuery(entry, query) {
			results = append(results, entry)
		}
	}

	// Search long-term memory if requested
	if searchLongTerm {
		for _, entry := range mm.longTerm {
			if mm.entryMatchesQuery(entry, query) {
				results = append(results, entry)
			}
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

// entryMatchesQuery checks if an entry matches the search query
func (mm *MemoryManager) entryMatchesQuery(entry *MemoryEntry, query string) bool {
	// Check key
	if strings.Contains(strings.ToLower(entry.Key), query) {
		return true
	}

	// Check value (convert to string for searching)
	valueStr := fmt.Sprintf("%v", entry.Value)
	if strings.Contains(strings.ToLower(valueStr), query) {
		return true
	}

	return false
}

// GetMemoryStats returns memory usage statistics
func (mm *MemoryManager) GetMemoryStats() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Count non-expired short-term entries
	validShortTerm := 0
	now := time.Now()
	for _, entry := range mm.shortTerm {
		if entry.TTL == nil || now.Sub(entry.Timestamp) <= *entry.TTL {
			validShortTerm++
		}
	}

	return map[string]interface{}{
		"short_term_count":      validShortTerm,
		"short_term_capacity":   mm.config.ShortTermCapacity,
		"long_term_count":       len(mm.longTerm),
		"short_term_percentage": float64(validShortTerm) / float64(mm.config.ShortTermCapacity) * 100,
	}
}