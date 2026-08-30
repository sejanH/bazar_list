package storage_test

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*storage.MySQLStorage, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.ShoppingList{},
		&models.Item{},
		&models.ListMember{},
		&models.ListActivity{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate tables: %v", err)
	}

	store := storage.NewStorageWithDB(db)
	return store, db
}

func TestPermissionsAndSharing(t *testing.T) {
	store, db := setupTestDB(t)

	// Create users
	user1 := models.User{Name: "Alice", Email: "alice@example.com", Password: "hash"}
	user2 := models.User{Name: "Bob", Email: "bob@example.com", Password: "hash"}
	user3 := models.User{Name: "Charlie", Email: "charlie@example.com", Password: "hash"}
	db.Create(&user1)
	db.Create(&user2)
	db.Create(&user3)

	// Create a shopping list owned by user1
	list := models.ShoppingList{
		UserID: user1.ID,
		Name:   "Weekly Groceries",
		Date:   time.Now(),
	}
	db.Create(&list)

	// 1. Check IsListOwner
	isOwner, err := store.IsListOwner(user1.ID, list.ID)
	if err != nil || !isOwner {
		t.Fatalf("expected user1 to be owner, got isOwner=%v, err=%v", isOwner, err)
	}
	isOwner, err = store.IsListOwner(user2.ID, list.ID)
	if err != nil || isOwner {
		t.Fatalf("expected user2 not to be owner, got isOwner=%v, err=%v", isOwner, err)
	}

	// 2. Check CanViewList before adding member
	canView, err := store.CanViewList(user1.ID, list.ID)
	if err != nil || !canView {
		t.Fatalf("expected owner user1 to view list, got %v, %v", canView, err)
	}
	canView, err = store.CanViewList(user2.ID, list.ID)
	if err != nil || canView {
		t.Fatalf("expected non-member user2 not to view list, got %v, %v", canView, err)
	}

	// 3. Add user2 as Editor
	err = store.AddListMember(list.ID, user2.ID, models.RoleEditor)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	// Add user3 as Viewer
	err = store.AddListMember(list.ID, user3.ID, models.RoleViewer)
	if err != nil {
		t.Fatalf("failed to add member user3: %v", err)
	}

	// 4. Verify CanViewList for members
	canView, err = store.CanViewList(user2.ID, list.ID)
	if err != nil || !canView {
		t.Fatalf("expected member user2 to view list, got %v, %v", canView, err)
	}
	canView, err = store.CanViewList(user3.ID, list.ID)
	if err != nil || !canView {
		t.Fatalf("expected member user3 to view list, got %v, %v", canView, err)
	}

	// 5. Verify CanEditList
	canEdit, err := store.CanEditList(user1.ID, list.ID) // Owner
	if err != nil || !canEdit {
		t.Fatalf("expected owner user1 to edit list, got %v, %v", canEdit, err)
	}

	canEdit, err = store.CanEditList(user2.ID, list.ID) // Editor
	if err != nil || !canEdit {
		t.Fatalf("expected editor user2 to edit list, got %v, %v", canEdit, err)
	}

	canEdit, err = store.CanEditList(user3.ID, list.ID) // Viewer
	if err != nil || canEdit {
		t.Fatalf("expected viewer user3 NOT to edit list, got %v, %v", canEdit, err)
	}

	// 6. Test UpdateMemberRole (promote user3 to editor)
	err = store.UpdateMemberRole(list.ID, user3.ID, models.RoleEditor)
	if err != nil {
		t.Fatalf("failed to update member role: %v", err)
	}
	canEdit, err = store.CanEditList(user3.ID, list.ID)
	if err != nil || !canEdit {
		t.Fatalf("expected promoted user3 to edit list, got %v, %v", canEdit, err)
	}

	// 7. Test GetListMembers
	members, err := store.GetListMembers(list.ID)
	if err != nil {
		t.Fatalf("failed to get list members: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].User == nil || members[1].User == nil {
		t.Fatalf("expected preloaded User on members")
	}

	// 8. Test RemoveListMember
	err = store.RemoveListMember(list.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to remove member: %v", err)
	}
	canView, err = store.CanViewList(user2.ID, list.ID)
	if err != nil || canView {
		t.Fatalf("expected removed member user2 not to view list, got %v, %v", canView, err)
	}

	// 9. Test GetListsByUserID includes shared lists
	user3Lists, err := store.GetListsByUserID(user3.ID)
	if err != nil {
		t.Fatalf("failed to get lists for user3: %v", err)
	}
	if len(user3Lists) != 1 || user3Lists[0].ID != list.ID {
		t.Fatalf("expected user3 to see shared list #%d, got %v", list.ID, user3Lists)
	}
}

func TestActivityLogs(t *testing.T) {
	store, _ := setupTestDB(t)

	activity1 := models.ListActivity{
		ListID:    10,
		UserID:    1,
		UserName:  "Alice",
		Action:    models.ActionItemAdded,
		ItemName:  "Milk",
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}
	activity2 := models.ListActivity{
		ListID:    10,
		UserID:    2,
		UserName:  "Bob",
		Action:    models.ActionItemPurchased,
		ItemName:  "Milk",
		CreatedAt: time.Now().Add(-1 * time.Minute),
	}
	activity3 := models.ListActivity{
		ListID:    10,
		UserID:    1,
		UserName:  "Alice",
		Action:    models.ActionItemDeleted,
		ItemName:  "Bread",
		CreatedAt: time.Now(),
	}

	if err := store.LogActivity(&activity1); err != nil {
		t.Fatalf("failed to log activity 1: %v", err)
	}
	if err := store.LogActivity(&activity2); err != nil {
		t.Fatalf("failed to log activity 2: %v", err)
	}
	if err := store.LogActivity(&activity3); err != nil {
		t.Fatalf("failed to log activity 3: %v", err)
	}

	activities, err := store.GetListActivities(10, 2)
	if err != nil {
		t.Fatalf("failed to get list activities: %v", err)
	}

	if len(activities) != 2 {
		t.Fatalf("expected 2 activities with limit=2, got %d", len(activities))
	}

	// Check order is descending
	if activities[0].Action != models.ActionItemDeleted {
		t.Fatalf("expected newest activity first (ActionItemDeleted), got %s", activities[0].Action)
	}
	if activities[1].Action != models.ActionItemPurchased {
		t.Fatalf("expected second newest activity (ActionItemPurchased), got %s", activities[1].Action)
	}
}
