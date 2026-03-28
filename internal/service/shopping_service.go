package service

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
	"github.com/sejan/bazarlist/pkg/logger"
)

// ShoppingService handles business logic for the shopping list
type ShoppingService struct {
	storage *storage.JSONStorage
	list    *models.ShoppingList
	logger  *logger.Logger
}

// NewShoppingService creates a new shopping service
func NewShoppingService(dataDir string) (*ShoppingService, error) {
	log := logger.NewLogger()

	// Create storage
	store, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Load existing list or create new one
	list, err := store.Load()
	if err != nil {
		log.Error("Failed to load shopping list, creating new one: %v", err)
		list = models.NewShoppingList()
	}

	return &ShoppingService{
		storage: store,
		list:    list,
		logger:  log,
	}, nil
}

// AddItem adds a new item to the list
func (s *ShoppingService) AddItem(name string, category models.Category) (*models.Item, error) {
	if name == "" {
		return nil, fmt.Errorf("item name cannot be empty")
	}

	// Set default category if not provided
	if category == "" {
		category = models.CategoryOther
	}

	item := s.list.AddItem(name, category)
	if err := s.save(); err != nil {
		return nil, err
	}

	s.logger.Info("Added item: %s (category: %s)", name, category)
	return item, nil
}

// RemoveItem removes an item from the list
func (s *ShoppingService) RemoveItem(id int) error {
	item := s.list.GetItem(id)
	if item == nil {
		return fmt.Errorf("item with ID %d not found", id)
	}

	if !s.list.RemoveItem(id) {
		return fmt.Errorf("failed to remove item with ID %d", id)
	}

	if err := s.save(); err != nil {
		return err
	}

	s.logger.Info("Removed item: %s (ID: %d)", item.Name, id)
	return nil
}

// CompleteItem marks an item as completed
func (s *ShoppingService) CompleteItem(id int) error {
	item := s.list.GetItem(id)
	if item == nil {
		return fmt.Errorf("item with ID %d not found", id)
	}

	if item.IsCompleted() {
		return fmt.Errorf("item with ID %d is already completed", id)
	}

	item.MarkCompleted()
	if err := s.save(); err != nil {
		return err
	}

	s.logger.Info("Completed item: %s (ID: %d)", item.Name, id)
	return nil
}

// GetAllItems returns all items, sorted by creation date
func (s *ShoppingService) GetAllItems() []*models.Item {
	items := make([]*models.Item, len(s.list.Items))
	copy(items, s.list.Items)

	// Sort by creation date (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return items
}

// GetPendingItems returns all pending items
func (s *ShoppingService) GetPendingItems() []*models.Item {
	return s.list.GetPendingItems()
}

// GetCompletedItems returns all completed items
func (s *ShoppingService) GetCompletedItems() []*models.Item {
	return s.list.GetCompletedItems()
}

// SearchItems searches for items by name
func (s *ShoppingService) SearchItems(term string) []*models.Item {
	return s.list.SearchItems(term)
}

// GetItemsByCategory returns items filtered by category
func (s *ShoppingService) GetItemsByCategory(category models.Category) []*models.Item {
	return s.list.GetItemsByCategory(category)
}

// GetItemCount returns the total number of items
func (s *ShoppingService) GetItemCount() int {
	return s.list.Count()
}

// save persists the current state
func (s *ShoppingService) save() error {
	if err := s.storage.Save(s.list); err != nil {
		s.logger.Error("Failed to save shopping list: %v", err)
		return fmt.Errorf("failed to save: %w", err)
	}
	return nil
}

// GetStoragePath returns the path to the storage file
func (s *ShoppingService) GetStoragePath() string {
	return s.storage.GetFilePath()
}

// GetRelativeStoragePath returns the storage path relative to the current directory
func (s *ShoppingService) GetRelativeStoragePath() (string, error) {
	absPath, err := filepath.Abs(s.storage.GetFilePath())
	if err != nil {
		return "", err
	}
	return absPath, nil
}
