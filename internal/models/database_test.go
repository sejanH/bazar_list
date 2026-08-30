package models_test

import (
	"testing"
	"time"

	"github.com/sejan/bazarlist/internal/models"
)

func TestListMemberAndActivityModels(t *testing.T) {
	now := time.Now()
	member := models.ListMember{
		ID:        1,
		ListID:    10,
		UserID:    20,
		Role:      models.RoleEditor,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if member.Role != "editor" {
		t.Fatalf("expected role editor, got %s", member.Role)
	}
	if member.ListID != 10 || member.UserID != 20 {
		t.Fatalf("expected list 10 and user 20, got list %d user %d", member.ListID, member.UserID)
	}

	// Verify roles constants
	if models.RoleOwner != "owner" || models.RoleEditor != "editor" || models.RoleViewer != "viewer" {
		t.Fatalf("unexpected role constants")
	}

	activity := models.ListActivity{
		ID:        1,
		ListID:    10,
		UserID:    20,
		UserName:  "Sarah",
		Action:    models.ActionItemPurchased,
		ItemName:  "Eggs",
		Details:   "Marked as purchased",
		CreatedAt: now,
	}

	if activity.Action != "item_purchased" {
		t.Fatalf("expected action item_purchased, got %s", activity.Action)
	}

	// Verify action constants
	expectedActions := map[string]string{
		models.ActionItemAdded:     "item_added",
		models.ActionItemUpdated:   "item_updated",
		models.ActionItemPurchased: "item_purchased",
		models.ActionItemDeleted:   "item_deleted",
		models.ActionMemberJoined:  "member_joined",
		models.ActionMemberRemoved: "member_removed",
	}

	for constant, val := range expectedActions {
		if constant != val {
			t.Fatalf("action mismatch: %s != %s", constant, val)
		}
	}

	// Verify request structs
	inviteReq := models.InviteMemberRequest{
		Email: "test@example.com",
		Role:  models.RoleEditor,
	}
	if inviteReq.Email != "test@example.com" || inviteReq.Role != models.RoleEditor {
		t.Fatalf("unexpected invite request values")
	}

	updateReq := models.UpdateMemberRoleRequest{
		Role: models.RoleViewer,
	}
	if updateReq.Role != models.RoleViewer {
		t.Fatalf("unexpected update role request value")
	}
}
