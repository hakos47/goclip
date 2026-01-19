package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ItemType defines the content type of a clipboard item.
type ItemType string

const (
	TypeText  ItemType = "text"
	TypeImage ItemType = "image"
)

// Item represents a single entry in the clipboard history.
type Item struct {
	Type      ItemType  `json:"type"`
	Content   string    `json:"content"` // Raw text or file path to image
	Preview   string    `json:"preview"` // Truncated text for UI display
	Timestamp time.Time `json:"timestamp"`
}

// History manages the storage and retrieval of clipboard items.
// It is safe for concurrent use.
type History struct {
	Items    []Item `json:"items"`
	mu       sync.RWMutex
	filePath string
	maxItems int
}

// NewHistory initializes the history manager.
func NewHistory(maxItems int) (*History, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "goclip")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app dir: %w", err)
	}

	h := &History{
		Items:    make([]Item, 0),
		filePath: filepath.Join(appDir, "history.json"),
		maxItems: maxItems,
	}

	h.load()
	return h, nil
}

// Add inserts a new item to the history.
// It handles deduplication and enforces the maximum item limit.
func (h *History) Add(item Item) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Deduplicate: Check if the new item is identical to the most recent one.
	if len(h.Items) > 0 {
		last := h.Items[0]
		if last.Type == item.Type && last.Content == item.Content {
			return nil // Duplicate ignored
		}
	}

	// Prepend new item
	h.Items = append([]Item{item}, h.Items...)

	// Prune old items
	if len(h.Items) > h.maxItems {
		removed := h.Items[h.maxItems]
		// Clean up image resources if necessary
		if removed.Type == TypeImage {
			_ = os.Remove(removed.Content)
		}
		h.Items = h.Items[:h.maxItems]
	}

	return h.save()
}

// GetItems returns a copy of the current items.
func (h *History) GetItems() []Item {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Return a copy to avoid race conditions on the slice
	result := make([]Item, len(h.Items))
	copy(result, h.Items)
	return result
}

// load reads the history from disk.
func (h *History) load() {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return // File might not exist yet, which is fine
	}

	_ = json.Unmarshal(data, &h.Items)
}

// save writes the history to disk atomically.
func (h *History) save() error {
	data, err := json.MarshalIndent(h.Items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.filePath, data, 0644)
}

// SaveImage writes image bytes to a persistent file and returns its path.
func SaveImage(data []byte) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	imgDir := filepath.Join(configDir, "goclip", "images")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("img_%d.png", time.Now().UnixNano())
	path := filepath.Join(imgDir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return path, nil
}