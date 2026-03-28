package models

import (
	"time"
)

// Category represents the type of item
type Category string

const (
	CategoryProduce    Category = "produce"
	CategoryDairy      Category = "dairy"
	CategoryMeat       Category = "meat"
	CategoryPantry     Category = "pantry"
	CategoryFrozen     Category = "frozen"
	CategoryBakery     Category = "bakery"
	CategoryBeverages  Category = "beverages"
	CategoryHousehold  Category = "household"
	CategoryOther      Category = "other"
)

// Status represents the completion status of an item
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
)

// Item represents a shopping list item
type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Category  Category  `json:"category"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewItem creates a new Item with the given name and category
func NewItem(id int, name string, category Category) *Item {
	now := time.Now()
	return &Item{
		ID:        id,
		Name:      name,
		Category:  category,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkCompleted marks the item as completed
func (i *Item) MarkCompleted() {
	i.Status = StatusCompleted
	i.UpdatedAt = time.Now()
}

// MarkPending marks the item as pending
func (i *Item) MarkPending() {
	i.Status = StatusPending
	i.UpdatedAt = time.Now()
}

// IsValid checks if the item has valid data
func (i *Item) IsValid() bool {
	return i.ID > 0 && i.Name != ""
}

// GetCategory returns the category of the item
func (i *Item) GetCategory() Category {
	return i.Category
}

// IsCompleted returns true if the item is completed
func (i *Item) IsCompleted() bool {
	return i.Status == StatusCompleted
}

// ShoppingList represents a collection of items
type ShoppingList struct {
	Items  []*Item `json:"items"`
	NextID int     `json:"next_id"`
}

// NewShoppingList creates a new empty shopping list
func NewShoppingList() *ShoppingList {
	return &ShoppingList{
		Items:  make([]*Item, 0),
		NextID: 1,
	}
}

// AddItem adds a new item to the list
func (sl *ShoppingList) AddItem(name string, category Category) *Item {
	item := NewItem(sl.NextID, name, category)
	sl.Items = append(sl.Items, item)
	sl.NextID++
	return item
}

// RemoveItem removes an item by ID
func (sl *ShoppingList) RemoveItem(id int) bool {
	for i, item := range sl.Items {
		if item.ID == id {
			sl.Items = append(sl.Items[:i], sl.Items[i+1:]...)
			return true
		}
	}
	return false
}

// GetItem retrieves an item by ID
func (sl *ShoppingList) GetItem(id int) *Item {
	for _, item := range sl.Items {
		if item.ID == id {
			return item
		}
	}
	return nil
}

// GetPendingItems returns all pending items
func (sl *ShoppingList) GetPendingItems() []*Item {
	var pending []*Item
	for _, item := range sl.Items {
		if !item.IsCompleted() {
			pending = append(pending, item)
		}
	}
	return pending
}

// GetCompletedItems returns all completed items
func (sl *ShoppingList) GetCompletedItems() []*Item {
	var completed []*Item
	for _, item := range sl.Items {
		if item.IsCompleted() {
			completed = append(completed, item)
		}
	}
	return completed
}

// SearchItems searches for items by name
func (sl *ShoppingList) SearchItems(term string) []*Item {
	var results []*Item
	for _, item := range sl.Items {
		if contains(item.Name, term) {
			results = append(results, item)
		}
	}
	return results
}

// GetItemsByCategory returns items filtered by category
func (sl *ShoppingList) GetItemsByCategory(category Category) []*Item {
	var items []*Item
	for _, item := range sl.Items {
		if item.Category == category {
			items = append(items, item)
		}
	}
	return items
}

// Count returns the total number of items
func (sl *ShoppingList) Count() int {
	return len(sl.Items)
}

// contains is a helper function for case-insensitive search
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
