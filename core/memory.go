package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryType represents different types of memory
type MemoryType string

const (
	MemoryTypeShortTerm MemoryType = "short_term" // Recent conversations and context
	MemoryTypeLongTerm  MemoryType = "long_term"  // Important facts, preferences, and knowledge
	MemoryTypeEpisodic  MemoryType = "episodic"   // Specific events and experiences
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID        string                 `json:"id"`
	Type      MemoryType             `json:"type"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"` // For short-term memory
}

// MemoryStore interface defines the memory storage operations
type MemoryStore interface {
	Add(ctx context.Context, entry *MemoryEntry) error
	Get(ctx context.Context, id string) (*MemoryEntry, error)
	Search(ctx context.Context, query string, memoryType MemoryType, limit int) ([]*MemoryEntry, error)
	Delete(ctx context.Context, id string) error
	ListByType(ctx context.Context, memoryType MemoryType, limit int) ([]*MemoryEntry, error)
	CleanupExpired(ctx context.Context) error
}

// FileMemoryStore implements MemoryStore using file system
type FileMemoryStore struct {
	logger    *logrus.Logger
	dataDir   string
	mutex     sync.RWMutex
	indexFile string
}

// NewFileMemoryStore creates a new file-based memory store
func NewFileMemoryStore(logger *logrus.Logger, dataDir string) (*FileMemoryStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	store := &FileMemoryStore{
		logger:    logger,
		dataDir:   dataDir,
		indexFile: filepath.Join(dataDir, "memory_index.json"),
	}

	// Initialize index file if it doesn't exist
	if _, err := os.Stat(store.indexFile); errors.Is(err, os.ErrNotExist) {
		index := make(map[string]*MemoryEntry)
		if err := store.saveIndex(index); err != nil {
			return nil, fmt.Errorf("failed to initialize memory index: %w", err)
		}
	}

	return store, nil
}

// Add adds a new memory entry
func (fms *FileMemoryStore) Add(ctx context.Context, entry *MemoryEntry) error {
	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	fms.mutex.Lock()
	defer fms.mutex.Unlock()

	// Load current index
	index, err := fms.loadIndex()
	if err != nil {
		return fmt.Errorf("failed to load memory index: %w", err)
	}

	// Save memory entry to file
	entryFile := filepath.Join(fms.dataDir, entry.ID+".json")
	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory entry: %w", err)
	}

	if err := os.WriteFile(entryFile, entryData, 0644); err != nil {
		return fmt.Errorf("failed to write memory entry file: %w", err)
	}

	// Update index
	index[entry.ID] = entry
	if err := fms.saveIndex(index); err != nil {
		return fmt.Errorf("failed to save memory index: %w", err)
	}

	fms.logger.Debugf("Added memory entry %s of type %s", entry.ID, entry.Type)
	return nil
}

// Get retrieves a memory entry by ID
func (fms *FileMemoryStore) Get(ctx context.Context, id string) (*MemoryEntry, error) {
	fms.mutex.RLock()
	defer fms.mutex.RUnlock()

	index, err := fms.loadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load memory index: %w", err)
	}

	entry, exists := index[id]
	if !exists {
		return nil, fmt.Errorf("memory entry not found: %s", id)
	}

	// Check if expired (for short-term memory)
	if entry.Type == MemoryTypeShortTerm && entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		return nil, fmt.Errorf("memory entry expired: %s", id)
	}

	return entry, nil
}

// Search searches for memory entries by content (simple substring search for now)
func (fms *FileMemoryStore) Search(ctx context.Context, query string, memoryType MemoryType, limit int) ([]*MemoryEntry, error) {
	fms.mutex.RLock()
	defer fms.mutex.RUnlock()

	index, err := fms.loadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load memory index: %w", err)
	}

	var results []*MemoryEntry
	for _, entry := range index {
		// Filter by memory type if specified
		if memoryType != "" && entry.Type != memoryType {
			continue
		}

		// Check if expired
		if entry.Type == MemoryTypeShortTerm && entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
			continue
		}

		// Simple substring search
		if query == "" || containsIgnoreCase(entry.Content, query) {
			results = append(results, entry)
		}
	}

	// Sort by timestamp (newest first)
	results = sortMemoryEntriesByTime(results, true)

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Delete removes a memory entry
func (fms *FileMemoryStore) Delete(ctx context.Context, id string) error {
	fms.mutex.Lock()
	defer fms.mutex.Unlock()

	index, err := fms.loadIndex()
	if err != nil {
		return fmt.Errorf("failed to load memory index: %w", err)
	}

	if _, exists := index[id]; !exists {
		return fmt.Errorf("memory entry not found: %s", id)
	}

	// Remove from index
	delete(index, id)
	if err := fms.saveIndex(index); err != nil {
		return fmt.Errorf("failed to save memory index: %w", err)
	}

	// Remove file
	entryFile := filepath.Join(fms.dataDir, id+".json")
	if err := os.Remove(entryFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		fms.logger.Warnf("Failed to remove memory entry file %s: %v", entryFile, err)
	}

	fms.logger.Debugf("Deleted memory entry %s", id)
	return nil
}

// ListByType lists memory entries by type
func (fms *FileMemoryStore) ListByType(ctx context.Context, memoryType MemoryType, limit int) ([]*MemoryEntry, error) {
	return fms.Search(ctx, "", memoryType, limit)
}

// CleanupExpired removes expired short-term memory entries
func (fms *FileMemoryStore) CleanupExpired(ctx context.Context) error {
	fms.mutex.Lock()
	defer fms.mutex.Unlock()

	index, err := fms.loadIndex()
	if err != nil {
		return fmt.Errorf("failed to load memory index: %w", err)
	}

	now := time.Now()
	var expiredIDs []string

	for id, entry := range index {
		if entry.Type == MemoryTypeShortTerm && entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	if len(expiredIDs) == 0 {
		return nil
	}

	// Remove expired entries from index
	for _, id := range expiredIDs {
		delete(index, id)
		entryFile := filepath.Join(fms.dataDir, id+".json")
		if err := os.Remove(entryFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			fms.logger.Warnf("Failed to remove expired memory entry file %s: %v", entryFile, err)
		}
	}

	if err := fms.saveIndex(index); err != nil {
		return fmt.Errorf("failed to save memory index after cleanup: %w", err)
	}

	fms.logger.Infof("Cleaned up %d expired memory entries", len(expiredIDs))
	return nil
}

// Helper functions
func (fms *FileMemoryStore) loadIndex() (map[string]*MemoryEntry, error) {
	data, err := os.ReadFile(fms.indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory index file: %w", err)
	}

	var index map[string]*MemoryEntry
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal memory index: %w", err)
	}

	return index, nil
}

func (fms *FileMemoryStore) saveIndex(index map[string]*MemoryEntry) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory index: %w", err)
	}

	if err := os.WriteFile(fms.indexFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory index file: %w", err)
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func sortMemoryEntriesByTime(entries []*MemoryEntry, reverse bool) []*MemoryEntry {
	// Simple bubble sort for small datasets
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			shouldSwap := false
			if reverse {
				shouldSwap = entries[j].Timestamp.Before(entries[j+1].Timestamp)
			} else {
				shouldSwap = entries[j].Timestamp.After(entries[j+1].Timestamp)
			}

			if shouldSwap {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}
	return entries
}

import "strings"