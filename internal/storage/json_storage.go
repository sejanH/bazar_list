package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sejan/bazarlist/internal/models"
)

// JSONStorage handles JSON file-based storage
type JSONStorage struct {
	filePath string
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(dataDir, "shopping_list.json")
	return &JSONStorage{
		filePath: filePath,
	}, nil
}

// Save saves the shopping list to a JSON file
func (s *JSONStorage) Save(list *models.ShoppingList) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create backup of existing file
	if _, err := os.Stat(s.filePath); err == nil {
		backupPath := s.filePath + ".bak"
		if err := os.Rename(s.filePath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal shopping list: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load loads the shopping list from a JSON file
func (s *JSONStorage) Load() (*models.ShoppingList, error) {
	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Return empty list if file doesn't exist
		return models.NewShoppingList(), nil
	}

	// Read file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON
	var list models.ShoppingList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shopping list: %w", err)
	}

	// Handle empty items slice
	if list.Items == nil {
		list.Items = make([]*models.Item, 0)
	}

	return &list, nil
}

// GetFilePath returns the current file path
func (s *JSONStorage) GetFilePath() string {
	return s.filePath
}
