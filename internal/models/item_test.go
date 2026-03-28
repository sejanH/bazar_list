package models

import (
	"testing"
	"time"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		itemName string
		category Category
		wantErr  bool
	}{
		{
			name:     "valid item",
			id:       1,
			itemName: "Milk",
			category: CategoryDairy,
			wantErr:  false,
		},
		{
			name:     "empty name",
			id:       1,
			itemName: "",
			category: CategoryDairy,
			wantErr:  false, // NewItem doesn't validate, it just creates
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.id, tt.itemName, tt.category)

			if item.ID != tt.id {
				t.Errorf("NewItem() ID = %v, want %v", item.ID, tt.id)
			}

			if item.Name != tt.itemName {
				t.Errorf("NewItem() Name = %v, want %v", item.Name, tt.itemName)
			}

			if item.Category != tt.category {
				t.Errorf("NewItem() Category = %v, want %v", item.Category, tt.category)
			}

			if item.Status != StatusPending {
				t.Errorf("NewItem() Status = %v, want %v", item.Status, StatusPending)
			}

			if time.Since(item.CreatedAt) > time.Second {
				t.Errorf("NewItem() CreatedAt seems too old")
			}
		})
	}
}

func TestItem_MarkCompleted(t *testing.T) {
	item := NewItem(1, "Milk", CategoryDairy)

	// Initial state should be pending
	if item.Status != StatusPending {
		t.Errorf("Initial status should be pending, got %v", item.Status)
	}

	// Mark as completed
	item.MarkCompleted()

	if item.Status != StatusCompleted {
		t.Errorf("Status after MarkCompleted() = %v, want %v", item.Status, StatusCompleted)
	}

	if !item.IsCompleted() {
		t.Error("IsCompleted() should return true after MarkCompleted()")
	}
}

func TestItem_MarkPending(t *testing.T) {
	item := NewItem(1, "Milk", CategoryDairy)
	item.MarkCompleted()

	// Mark as pending
	item.MarkPending()

	if item.Status != StatusPending {
		t.Errorf("Status after MarkPending() = %v, want %v", item.Status, StatusPending)
	}

	if item.IsCompleted() {
		t.Error("IsCompleted() should return false after MarkPending()")
	}
}

func TestItem_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		item  *Item
		want  bool
	}{
		{
			name: "valid item",
			item: NewItem(1, "Milk", CategoryDairy),
			want: true,
		},
		{
			name: "zero ID",
			item: &Item{ID: 0, Name: "Milk"},
			want: false,
		},
		{
			name: "negative ID",
			item: &Item{ID: -1, Name: "Milk"},
			want: false,
		},
		{
			name:  "empty name",
			item:  NewItem(1, "", CategoryDairy),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsValid(); got != tt.want {
				t.Errorf("Item.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewShoppingList(t *testing.T) {
	list := NewShoppingList()

	if list == nil {
		t.Fatal("NewShoppingList() returned nil")
	}

	if list.Items == nil {
		t.Error("Items slice should be initialized")
	}

	if len(list.Items) != 0 {
		t.Errorf("New list should have 0 items, got %d", len(list.Items))
	}

	if list.NextID != 1 {
		t.Errorf("NextID should start at 1, got %d", list.NextID)
	}
}

func TestShoppingList_AddItem(t *testing.T) {
	list := NewShoppingList()

	// Add first item
	item1 := list.AddItem("Milk", CategoryDairy)
	if item1.ID != 1 {
		t.Errorf("First item ID should be 1, got %d", item1.ID)
	}

	if len(list.Items) != 1 {
		t.Errorf("List should have 1 item, got %d", len(list.Items))
	}

	if list.NextID != 2 {
		t.Errorf("NextID should be 2, got %d", list.NextID)
	}

	// Add second item
	item2 := list.AddItem("Bread", CategoryBakery)
	if item2.ID != 2 {
		t.Errorf("Second item ID should be 2, got %d", item2.ID)
	}

	if len(list.Items) != 2 {
		t.Errorf("List should have 2 items, got %d", len(list.Items))
	}
}

func TestShoppingList_RemoveItem(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)
	list.AddItem("Apples", CategoryProduce)

	// Remove middle item
	removed := list.RemoveItem(2)
	if !removed {
		t.Error("RemoveItem() should return true for existing item")
	}

	if len(list.Items) != 2 {
		t.Errorf("List should have 2 items after removal, got %d", len(list.Items))
	}

	// Check remaining items
	if list.Items[0].ID != 1 {
		t.Errorf("First item should still have ID 1, got %d", list.Items[0].ID)
	}

	if list.Items[1].ID != 3 {
		t.Errorf("Third item should still have ID 3, got %d", list.Items[1].ID)
	}

	// Try to remove non-existent item
	removed = list.RemoveItem(99)
	if removed {
		t.Error("RemoveItem() should return false for non-existent item")
	}
}

func TestShoppingList_GetItem(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)

	// Get existing item
	item := list.GetItem(1)
	if item == nil {
		t.Fatal("GetItem() should not return nil for existing item")
	}

	if item.Name != "Milk" {
		t.Errorf("Got item with name %s, want Milk", item.Name)
	}

	// Get non-existent item
	item = list.GetItem(99)
	if item != nil {
		t.Error("GetItem() should return nil for non-existent item")
	}
}

func TestShoppingList_GetPendingItems(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)

	// Mark one as completed
	list.GetItem(1).MarkCompleted()

	pending := list.GetPendingItems()
	if len(pending) != 1 {
		t.Errorf("Should have 1 pending item, got %d", len(pending))
	}

	if pending[0].ID != 2 {
		t.Errorf("Pending item should have ID 2, got %d", pending[0].ID)
	}
}

func TestShoppingList_GetCompletedItems(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)

	// Mark one as completed
	list.GetItem(1).MarkCompleted()

	completed := list.GetCompletedItems()
	if len(completed) != 1 {
		t.Errorf("Should have 1 completed item, got %d", len(completed))
	}

	if completed[0].ID != 1 {
		t.Errorf("Completed item should have ID 1, got %d", completed[0].ID)
	}
}

func TestShoppingList_SearchItems(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)
	list.AddItem("Butter", CategoryDairy)

	// Search for "mil"
	results := list.SearchItems("mil")
	if len(results) != 1 {
		t.Errorf("Search should find 1 item, got %d", len(results))
	}

	if results[0].Name != "Milk" {
		t.Errorf("Search result should be Milk, got %s", results[0].Name)
	}

	// Search for "ea" (should match Bread)
	results = list.SearchItems("ea")
	if len(results) != 1 {
		t.Errorf("Search should find 1 item, got %d", len(results))
	}

	// Search for non-existent term
	results = list.SearchItems("xyz")
	if len(results) != 0 {
		t.Errorf("Search should find 0 items, got %d", len(results))
	}
}

func TestShoppingList_GetItemsByCategory(t *testing.T) {
	list := NewShoppingList()
	list.AddItem("Milk", CategoryDairy)
	list.AddItem("Bread", CategoryBakery)
	list.AddItem("Cheese", CategoryDairy)

	dairy := list.GetItemsByCategory(CategoryDairy)
	if len(dairy) != 2 {
		t.Errorf("Should have 2 dairy items, got %d", len(dairy))
	}

	bakery := list.GetItemsByCategory(CategoryBakery)
	if len(bakery) != 1 {
		t.Errorf("Should have 1 bakery item, got %d", len(bakery))
	}

	meat := list.GetItemsByCategory(CategoryMeat)
	if len(meat) != 0 {
		t.Errorf("Should have 0 meat items, got %d", len(meat))
	}
}

func TestShoppingList_Count(t *testing.T) {
	list := NewShoppingList()

	if list.Count() != 0 {
		t.Errorf("Empty list should have count 0, got %d", list.Count())
	}

	list.AddItem("Milk", CategoryDairy)
	if list.Count() != 1 {
		t.Errorf("List with 1 item should have count 1, got %d", list.Count())
	}

	list.AddItem("Bread", CategoryBakery)
	if list.Count() != 2 {
		t.Errorf("List with 2 items should have count 2, got %d", list.Count())
	}
}
